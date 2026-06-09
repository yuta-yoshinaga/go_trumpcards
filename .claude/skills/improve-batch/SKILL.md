---
name: improve-batch
version: 1.0.0
description: Clear a whole batch of improvement/UX GitHub issues by running the single-issue loop repeatedly, lowest-effort first, one PR at a time until the range is done. Use for "issueバッチを片付けて", "#NNNN〜#MMMM を全部対応", "残りの改善issueを全部やって", bulk issue clearing.
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Grep
  - Glob
  - AskUserQuestion
# Explicit-invocation only (it commits, opens PRs, and merges in a loop); run it with `/improve-batch <range>`.
disable-model-invocation: true
---

# improve-batch — clear an issue batch, one merge at a time

Orchestrates [`improve-issue`](../improve-issue/SKILL.md) across a set of issues.
This is the loop that cleared **#2238–#2292** (20 PRs #2376–#2395 merged + 1
false-positive closed) in a single sustained run.

**This skill is the *orchestrator only*.** Every per-issue mechanic — branch,
TDD, PR, CI triage, full-review handling, squash-merge, sync — lives in
[`improve-issue`](../improve-issue/SKILL.md). Read that first; do not duplicate
its steps here.

**Input:** a scope — an inclusive number range (`/improve-batch 2238-2292`), a
label/milestone, or "the remaining open improvement issues". If ambiguous, ask.

## Why serial, not parallel

Process **one issue fully (through merge) before starting the next**:

- Merges serialize on `develop` anyway, and each new branch must start from the
  just-merged tip — so **re-sync `develop` between issues** or you branch stale.
- This box is RAM-constrained (~2 GB); parallel `bun`/`go`/builds thrash/OOM.
- Each issue's CI (esp. E2E ~15–20 min) is the real wall-clock cost; running
  many at once doesn't speed merges and multiplies flake reruns.

So the throughput pattern is: implement issue N → push → **while its CI runs,
scope issue N+1 read-only** (don't edit — you're still on N's branch) → merge N
→ branch N+1. One green PR in flight at a time.

## Loop

1. **Build the work-list.** Enumerate open issues in scope:
   ```sh
   for n in $(seq <lo> <hi>); do \
     s=$(gh issue view $n --json state,title -q '.state+"\t"+.title' 2>/dev/null); \
     echo "$s" | grep -q '^OPEN' && echo "$n $s"; done
   ```
   (or `gh issue list --label <l> --state open`).

2. **Order lowest-effort-first** (cheapest, highest-confidence first — builds
   momentum and front-loads easy wins). Rough rubric:
   1. **Close-as-false-positive** (premise wrong / already done) — seconds.
   2. **i18n-only / static-label** changes (no logic).
   3. **Single-page frontend UX** (badge / highlight / progress bar / settings).
   4. **Pure-util extraction** mirroring Go domain (+ util test).
   5. **Shared-component** changes (touch a component many games use).
   6. **Backend Go** (domain / presenter / CUI) — heaviest, CI-gated.
   Read each issue body to place it; reorder freely as you learn the code.

3. **For each issue, run `improve-issue`** end-to-end (branch → … → merge →
   sync). Do **not** advance until the current PR is merged (or the issue is
   closed as a false-positive).

4. **Maintain a running tally** in the batch memory file (reuse the existing
   `memory/project_issues_*.md` for this batch, or start a new
   `project_issues_<lo>_<hi>.md`, and link it from `MEMORY.md`): which issues are merged, which
   PR numbers, recurring gotchas surfaced this run. Update it as you go so a
   resumed run has context.

5. **Stop and surface** (don't push through) when:
   - the range is exhausted (report the final list of merged PRs + closed issues);
   - an issue genuinely needs a product/UX decision (use `AskUserQuestion`);
   - the same CI failure recurs after a flake rerun (it may be real — inspect with
     `gh run view --job <id> --log-failed`);
   - a change would touch architecture/ADR territory (out of "improvement" scope).

## Reporting

When the batch is done, post a single summary: a table of `issue → PR → game →
one-line change`, plus any issues closed as false-positives (with why) and any
deferred for a decision. (Mirror the wrap-up format used for #2238–#2292.)

## Pacing note

A full batch is a long autonomous run (each issue ≈ implement + ~20 min CI +
review fixups). It's fine to leave it running; it serializes safely and can be
resumed from the memory tally. The user can interrupt at any merged boundary.
