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
Works through the issues in lowest-effort order. Lanes may run concurrently; merges into develop land one at a time.
When processing issues concurrently in isolated worktrees, use the helper scripts in
[`batch-lanes`](../batch-lanes/SKILL.md) for lane setup, delegation, gates, and shipping.

**This skill is the *orchestrator only*.** Every per-issue mechanic — branch,
TDD, PR, CI triage, full-review handling, squash-merge, sync — lives in
[`improve-issue`](../improve-issue/SKILL.md). Read that first; do not duplicate
its steps here.

**Input:** a scope — an inclusive number range (`/improve-batch 2238-2292`), a
label/milestone, or "the remaining open improvement issues". If ambiguous, ask.

## What runs concurrently, what stays serial

Issues may be implemented concurrently in separate worktree lanes (the batch-lanes
scripts); what stays serial is landing them on `develop`:

- Every lane branches from freshly fetched `develop`. When `develop` moves under an
  open PR, update the branch (`gh pr update-branch`) before landing it.
- Land one PR at a time through `land-pr`, and after each merge run `go build ./...`
  on `develop`: two PRs can each add the same symbol, and each is green on its own.
- CI here runs one workflow at a time, so more open PRs queue CI rather than speed
  up merges. Keep only as many lanes open as you can review.

## Loop

1. **Build the work-list.** Enumerate open issues in scope:
   ```sh
   for n in $(seq <lo> <hi>); do \
     s=$(gh issue view $n --json state,title -q '.state+"\t"+.title' 2>/dev/null); \
     echo "$s" | grep -q '^OPEN' && echo "$n $s"; done
   ```
   (or `gh issue list --label <l> --state open`).

2. **Triage (read-only).** Before implementing anything, run one read-only pass
   over every issue in scope. Each scan returns `VERDICT: IMPLEMENT | NOT_PLANNED`
   with file:line evidence. Close the `NOT_PLANNED` issues in bulk first. The
   orchestrator must read the cited lines for every `NOT_PLANNED` verdict itself:
   last time delegated `NOT_PLANNED` verdicts were wrong in 4 of the first 7,
   rejecting whole proposals because one part of the premise was off. If the same
   premise recurs across 3 or more issues, stop and turn it into one policy
   question for the user rather than deciding it per issue.

3. **Order lowest-effort-first** (cheapest, highest-confidence first — builds
   momentum and front-loads easy wins). Rough rubric:
   1. **Close-as-false-positive** (premise wrong / already done) — seconds.
   2. **i18n-only / static-label** changes (no logic).
   3. **Single-page frontend UX** (badge / highlight / progress bar / settings).
   4. **Pure-util extraction** mirroring Go domain (+ util test).
   5. **Shared-component** changes (touch a component many games use).
   6. **Backend Go** (domain / presenter / CUI) — heaviest, CI-gated.
   Read each issue body to place it; reorder freely as you learn the code.

4. **For each issue, run `improve-issue`** end-to-end (branch → … → merge →
   sync). With lanes, start the next issue in a free lane once the current one is
   pushed; land PRs in the order they turn green.

5. **Maintain a running tally** in the batch memory file (reuse the existing
   `memory/project_issues_*.md` for this batch, or start a new
   `project_issues_<lo>_<hi>.md`, and link it from `MEMORY.md`): which issues are merged, which
   PR numbers, recurring gotchas surfaced this run. Update it as you go so a
   resumed run has context.

6. **Stop and surface** (don't push through) when:
   - the range is exhausted (report the final list of merged PRs + closed issues);
   - an issue genuinely needs a product/UX decision (use `AskUserQuestion`);
   - the same CI failure recurs after a flake rerun (it may be real — inspect with
     `gh run view --job <id> --log-failed`);
   - a change would touch architecture/ADR territory (out of "improvement" scope).

## Reporting

When the batch is done, post a single summary: a table of `issue → PR → game →
one-line change`, plus any issues closed as false-positives (with why) and any
deferred for a decision.

## Pacing note

A full batch is a long autonomous run (each issue ≈ implement + a full CI run +
review fixups). It's fine to leave it running; it serializes safely and can be
resumed from the memory tally. The user can interrupt at any merged boundary.
