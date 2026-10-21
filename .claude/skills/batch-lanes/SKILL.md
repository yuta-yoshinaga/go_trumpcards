---
name: batch-lanes
version: 1.0.0
description: Run improve-batch issues concurrently in isolated worktree lanes. Scripts used by Claude from improve-batch; not a user-facing entry point.
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Grep
  - Glob
---

# batch-lanes

Helper scripts for Claude while `improve-batch` delegates several issues into separate worktree slots. Invoke scripts by their paths under this skill's `scripts/` directory.

## Standard flow

1. `setup_wt.sh`: create or prepare `wt-ib<slot>` and branch it for an issue.
2. `go.sh`: delegate an issue using the exact prompt file supplied as argument 3 or `BATCH_PROMPT_FILE`.
3. `rcg.sh`: serialize and run the receipt gate for that slot, then wait with `waitdone.sh <slot>`.
4. `ship.sh`: commit a gated issue worktree and open its PR.
5. `triage.sh` / `sweep.sh`: review posted automated feedback and land reviewed, green PRs; after ship, use `drain.sh` to retry pending PRs every 5 minutes. It exits on FAIL, CONFLICT, REVIEW-READY, or all merged. This replaces hand-written `sleep` loops.
6. `fx.sh` and `gofx.sh`: prepare a fixup worktree and delegate a fix prompt.
7. `fixpush.sh`: push a gated fix and reply on the PR.
8. `drain.sh`: wait for the selected batch PRs to merge.
9. Before the closing PR of a batch, run `combined.sh` with the batch's guard specs to catch failures that appear only when slices are combined.

Set `BATCH_BRANCH_RE` to a regular expression matching the batch branches before running `triage.sh`, `sweep.sh`, or `drain.sh`. Set `BATCH_REPO` to the repository worktree, `BATCH_WT_ROOT` to the parent directory for `wt-ib<slot>` worktrees, or `BATCH_STATE` to override state storage. By default, worktrees are created beside the repository and all state files live in `<git-common-dir>/batch-lanes`, shared by linked worktrees and outside version control. Prompt files are passed by path; scripts do not infer a filename from an issue or PR number. Set `CLAUDE_SESSION_URL` to include a session reference in commit and PR metadata.

## Scripts

## Prerequisites

The flows require `dele`, `gh`, `jq`, `flock`, `bun`, `go`, and `golangci-lint`, in addition to Bash and Git. `dele` is used by `go.sh` and `gofx.sh`; if it is missing, the command fails, and the script records the nonzero exit status and log instead of reporting successful delegation. Install the tools used by the flow before running it.

- `common.sh`: resolve repository, worktree, and shared state paths and validate required environment.
- `setup_wt.sh`: fetch `develop` and prepare a clean issue worktree.
- `sync.sh <slot>`: merge `origin/develop` into a lane while preserving its changes as uncommitted work; after syncing, rerun `rcg.sh <slot>` before `ship.sh`. On conflict it restores the lane and prints the original ref plus a manual stash/fast-forward rebase command.
- `combined.sh`: merge matching open batch branches into a scratch build and repeat the supplied Playwright guard specs; this catches deal-dependent failures like the Mushi and Desmoche failures found only after combining slices.
- `go.sh`: run `dele` for a prompt file in an issue worktree and save a fresh per-run log; only a line starting with `VERDICT: NOT_PLANNED` suppresses `WARN NO_CHANGE`, which otherwise fails the delegation.
- `gofx.sh`: run `dele` for a prompt file in a fixup worktree after checking a PR branch; the slot must be numeric like `fx.sh`'s.
- `fx.sh`: check out an existing remote PR branch in a fixup worktree; use a numeric slot distinct from active lanes, such as 21..29.
- `stopsrv.sh`: stop server processes listening on lane ports by numeric slot (or all slots when omitted).
- `busy.sh`: report delegation processes writing in a worktree.
- `treehash.sh`: hash the worktree state for the gate stamp.
- `receipt.sh`: run relevant Vitest files plus frontend and backend checks for a worktree; route selection joins consecutive branch words and requires exact game-name matches; `WARN NO_TEST_MATCH` means branch or changed-game selection found no page test; it is not reported when no changed file is under `frontend/`, a `*Presenter.go`, or a Go locale, since no page renders such a change.
- `rc.sh`: save a concise receipt with test shrink and dead-code warnings.
- `overlap.sh`: report files changed by this lane and another active lane; set `ALLOW_OVERLAP="file ..."` to suppress known shared files.
- `adrcheck.sh <slot>`: warn for each changed `docs/adr/*.md`; set `ALLOW_ADR_EDIT=1` when the orchestrator writes the ADR progress line after sync and before ship.
- `rc.sh` also records `WARN WORKER_BUILD`, `WARN UNUSED_NEW_SYMBOL`, and `WARN PATCH_BRANCH_GAP`; set `ALLOW_WORKER_BUILD=1`, `ALLOW_UNUSED="name ..."`, or `ALLOW_PATCH_GAPS=1` to allow known exceptions.
- `identgone.sh`: report a removed identifier as `WARN IDENT_GONE <file> <old> -> <new>` when a newly appearing identifier in the same changed `.ts`, `.tsx`, or `.go` file contains it, is contained by it (both names at least 6 characters), or is within case-insensitive edit distance 2 (both names at least 6 characters), and both identifiers have the same occurrence count. When several candidates qualify, the closest edit distance wins, with name order breaking ties. Other removals are `NOTE IDENT_GONE` and do not fail the receipt. Comments, strings, and numeric literal fragments are ignored. Set `ALLOW_IDENT_GONE="name1 name2"` to suppress either report for known intentional removals; the names apply regardless of file.
- `rc.sh` also reruns the Go test functions the change adds or edits (`repeattests.sh`, `-count=50`) and records `WARN DEAL_DEPENDENT` when one fails, which usually means it depends on the deal. Set `REPEAT_TESTS_COUNT` to change the count, or `0` to skip.
- `rcg.sh`: serialize receipt gates with `flock` and write a fresh slot result. Before the receipt, `NOTE STALE_BASE <file> (run sync.sh <slot>)` reports files changed on develop since the lane branched.
- `rcg.sh` records its PID per slot and stops the previous gate process tree only when `/proc` confirms it is an `rcg.sh` for that slot; `mutate.sh <slot> <ref>:<path> <e2e args...>` prints its `backup: <path>` before checking, first verifies the unmutated lane passes the same E2E command, then temporarily swaps the file for a mutation check and restores it on exit. A failed baseline reports `ERROR baseline failed`; only a failure after mutation reports `DETECTED`.
- `e2e.sh`: build and run a lane's Playwright specs on its isolated port, with at most two lanes running concurrently.
- Set `E2E_SKIP_BUILD=1` to skip frontend/backend builds, server startup, and the port check; this assumes a server is already running on the lane's port.
- `e2egate.sh`: select E2E specs from a lane's frontend diff; deleted route references are repeated across sweep specs; `--dry-run` prints the selection without starting Playwright.
- `ship.sh`: commit a gated issue worktree and create a PR. It refuses files changed on `develop` since the lane branched; run `sync.sh <slot>`, then `rcg.sh <slot>` before shipping. Set `ALLOW_STALE_BASE=1` to bypass. It refuses a body that mentions other issue numbers but not `#n` (`ALLOW_UNMENTIONED_ISSUE=1`) or a branch whose `/<number>-` issue number differs from n (`ALLOW_BRANCH_ISSUE_MISMATCH=1`), and warns with `WARN: body closes #m` when closing keywords target another issue. The 4th argument is either summary lines (wrapped in `## Summary` / `Closes` / `## Test plan` / footer) or a complete body with its own `## ` headings, which is used as written. In that case only a missing `Closes #n` and a missing footer are added (`prbody.sh`).
- `fixpush.sh`: commit and push a gated fix and comment on its PR; it refuses when the PR is not OPEN or the worktree branch differs from the PR head. It appends the PR to `reviewed.txt` after the push and reply succeed, so do not add it by hand; if the reply fails after the push, check the PR and append it manually.
- `ship.sh` and `fixpush.sh` infer `Delegated-To` from `dele: <backend> ok` in this slot's `go-<slot>-*.log` files newer than the matching `slot<slot>.branch` marker written when the lane branch is created; a missing or mismatched marker reports `unknown`. `BATCH_DELEGATED_TO` overrides the inferred value, and no matching logs reports `none (orchestrator)`.
- `sweep.sh`: land eligible reviewed batch PRs and report blocked ones.
- `sweep.sh` reports `UNSWEPT` for open PRs outside `BATCH_BRANCH_RE`; set `BATCH_UNSWEPT_IGNORE="<pr ...>"` to ignore intentional non-batch PRs.
- `triage.sh`: inspect automated review feedback and mark clean reviews as read.
- `land.sh`: require the PR to be listed in `reviewed.txt`, require a passing `codecov/patch` result when the PR changes production Go or TypeScript, then run the merge gate and squash-merge. It refuses if a worktree on the PR branch has a running delegation, uncommitted changes, unpushed commits, or cannot be compared with origin (`ALLOW_LANE_PENDING=1`). Missing or pending coverage exits 3 to wait. An overlapping live gate prints `LANE-GATING` and exits 4; `ALLOW_LANE_GATING=1` bypasses this check. `BATCH_SKIP_CODECOV=1` bypasses the coverage check; unreviewed PRs are never bypassed.
- `waitfor.sh`: wait for a shell condition up to a timeout.
- `waitdone.sh`: wait for a slot's `rcg` result written after its latest start.
- `drain.sh`: repeatedly sweep the selected batch until it is clear or blocked.
- `fieldcheck.sh`: warn when each changed page reads a state field missing from its matching controller JSON output.
- `e2echeck.py`: find removed Japanese locale text still referenced by E2E specs.
- `rerun504.sh`: rerun completed PR jobs whose logs show an infrastructure HTTP 5xx.
- `mkprompt.sh`: print an issue delegation prompt with the shared game-page rules.

Shared issue delegation rules are in [`templates/tail.md`](templates/tail.md). The full issue prompt emitted by `mkprompt.sh` includes the game-page requirements.

## Known traps

- `rcg.sh` holds the shared state lock for the receipt, releases it before E2E, then checks the worktree hash and writes its per-slot gate stamp without reacquiring the lock.
- Gate success is tied to the current worktree hash. Any edits after `rcg.sh` require another gate before shipping.
- `go.sh` and `gofx.sh` require an existing, non-empty prompt path; an omitted or empty prompt never starts `dele`.
- Pass a prompt path directly. A prefix plus issue/PR tag can accidentally select a different or empty file.
- `waitdone.sh` uses fresh per-slot `rcg` result timestamps; old completion lines cannot satisfy a new wait.
- Every concurrent batch command must use the same `BATCH_STATE` and `BATCH_BRANCH_RE` so gates, review marks, and selection stay consistent.
