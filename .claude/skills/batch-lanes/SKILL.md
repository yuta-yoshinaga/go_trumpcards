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
3. `rcg.sh`: serialize and run the receipt gate for that slot.
4. `ship.sh`: commit a gated issue worktree and open its PR.
5. `triage.sh` / `sweep.sh`: review posted automated feedback and land reviewed, green PRs; `drain.sh` repeats the sweep.
6. `fx.sh` and `gofx.sh`: prepare a fixup worktree and delegate a fix prompt.
7. `fixpush.sh`: push a gated fix and reply on the PR.
8. `drain.sh`: wait for the selected batch PRs to merge.

Set `BATCH_BRANCH_RE` to a regular expression matching the batch branches before running `triage.sh`, `sweep.sh`, or `drain.sh`. Set `BATCH_REPO` to the repository worktree, `BATCH_WT_ROOT` to the parent directory for `wt-ib<slot>` worktrees, or `BATCH_STATE` to override state storage. By default, worktrees are created beside the repository and all state files live in `<git-common-dir>/batch-lanes`, shared by linked worktrees and outside version control. Prompt files are passed by path; scripts do not infer a filename from an issue or PR number. Set `CLAUDE_SESSION_URL` to include a session reference in commit and PR metadata.

## Scripts

## Prerequisites

The flows require `dele`, `gh`, `jq`, `flock`, `bun`, `go`, and `golangci-lint`, in addition to Bash and Git. `dele` is used by `go.sh` and `gofx.sh`; if it is missing, the command fails, and the script records the nonzero exit status and log instead of reporting successful delegation. Install the tools used by the flow before running it.

- `common.sh`: resolve repository, worktree, and shared state paths and validate required environment.
- `setup_wt.sh`: fetch `develop` and prepare a clean issue worktree.
- `go.sh`: run `dele` for a prompt file in an issue worktree and save a fresh per-run log.
- `gofx.sh`: run `dele` for a prompt file in a fixup worktree after checking a PR branch.
- `fx.sh`: check out an existing remote PR branch in a fixup worktree.
- `busy.sh`: report delegation processes writing in a worktree.
- `treehash.sh`: hash the worktree state for the gate stamp.
- `receipt.sh`: run relevant Vitest files plus frontend and backend checks for a worktree.
- `rc.sh`: save a concise receipt with test shrink and dead-code warnings.
- `rc.sh` also records `WARN WORKER_BUILD`, `WARN UNUSED_NEW_SYMBOL`, and `WARN PATCH_BRANCH_GAP`; set `ALLOW_WORKER_BUILD=1`, `ALLOW_UNUSED="name ..."`, or `ALLOW_PATCH_GAPS=1` to allow known exceptions.
- `rcg.sh`: serialize receipt gates with `flock` and write a fresh slot result.
- `ship.sh`: commit a gated issue worktree and create a PR.
- `fixpush.sh`: commit and push a gated fix and comment on its PR.
- `sweep.sh`: land eligible reviewed batch PRs and report blocked ones.
- `sweep.sh` reports `UNSWEPT` for open PRs outside `BATCH_BRANCH_RE`; set `BATCH_UNSWEPT_IGNORE="<pr ...>"` to ignore intentional non-batch PRs.
- `triage.sh`: inspect automated review feedback and mark clean reviews as read.
- `land.sh`: run the merge gate and squash-merge a PR.
- `waitfor.sh`: wait for a shell condition up to a timeout.
- `waitdone.sh`: wait for a slot's `rcg` result written after its latest start.
- `drain.sh`: repeatedly sweep the selected batch until it is clear or blocked.
- `fieldcheck.sh`: warn when page state fields have no controller JSON field.
- `e2echeck.py`: find removed Japanese locale text still referenced by E2E specs.
- `rerun504.sh`: rerun completed PR jobs whose logs show an infrastructure HTTP 5xx.
- `mkprompt.sh`: print an issue delegation prompt with the shared game-page rules.

Shared issue delegation rules are in [`templates/tail.md`](templates/tail.md). The full issue prompt emitted by `mkprompt.sh` includes the game-page requirements.

## Known traps

- A gate can start while another lane is running Vitest; `rcg.sh` holds a lock in the shared state directory so importer runs execute one at a time.
- Gate success is tied to the current worktree hash. Any edits after `rcg.sh` require another gate before shipping.
- `go.sh` and `gofx.sh` require an existing, non-empty prompt path; an omitted or empty prompt never starts `dele`.
- Pass a prompt path directly. A prefix plus issue/PR tag can accidentally select a different or empty file.
- `waitdone.sh` uses fresh per-slot `rcg` result timestamps; old completion lines cannot satisfy a new wait.
- Every concurrent batch command must use the same `BATCH_STATE` and `BATCH_BRANCH_RE` so gates, review marks, and selection stay consistent.
