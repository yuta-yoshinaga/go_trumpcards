---
name: land-pr
description: Verify the merge gate for a PR (head-SHA match, zero pending, zero failing, check-count floor, reviews read) and then squash-merge it. Use when a PR looks ready ("マージして", "land it", "merge #NNNN", "ready to merge?"). Refuses to merge when any condition fails.
# Model-invocable: gating this does not prevent merges, it only means they happen
# WITHOUT the gate -- the alternative is hand-rolling `gh pr merge` and re-deriving the
# head-SHA match, pending/failing counts and check-count floor by hand, which is strictly
# worse. The authorisation this relies on is written down in CLAUDE.md
# ("Autonomous Merge Policy"); do not infer it from this comment.
allowed-tools: Read, Grep, Glob, Bash
argument-hint: "<pr-number>"
---

# Land PR

Merging looks like one command. It is not, and the two ways it goes wrong both merged real PRs
early.

## Why this exists

**A watcher that waits for "no pending checks" lies right after a push.** The new run's checks
are not registered yet, so `gh pr checks` reports the *previous* commit's green result and the
loop exits immediately. Acting on that merged #4345 with 12 checks still pending. Pinning the
watcher to the CI *workflow* was not enough either — `gh pr checks` aggregates several workflows
(CodeQL, deploy jobs, codecov), so watching only `name=="CI"` reported success while
`Analyze (go, autobuild)` was still running, and #4347 merged early a second time.

The gate that actually holds needs **both** halves, every time:

```sh
[ "$(gh pr view N --json headRefOid --jq .headRefOid)" = "$(git rev-parse HEAD)" ]   # no newer push
[ "$(gh pr checks N --json bucket --jq '[.[]|select(.bucket=="pending")]|length')" = 0 ]
```

The head-match alone passes while checks are queued. The pending check alone passes against a
stale commit.

**Auto-review does not re-run on push.** A green tick after a fix-up commit means CI re-ran; it
does not mean anyone reviewed the fix-up. Read the review before merging, and know which commit
it was written against.

## Procedure

1. **Run the gate.** It decides nothing and merges nothing — it prints conditions.

   ```sh
   bash .claude/skills/land-pr/scripts/merge-gate.sh <pr-number>
   ```

   Pass the expected SHA explicitly if you are not on the PR branch:
   `merge-gate.sh <pr> <sha>`.

2. **Read every review.** The script lists bot comments and inline threads but cannot judge
   them. Open them:

   ```sh
   gh pr view <pr> --json comments --jq '.comments[].body'
   gh api repos/{owner}/{repo}/pulls/<pr>/comments --jq '.[]|"\(.path):\(.line)\n\(.body)\n---"'
   ```

   An inline comment whose `line` is `null` is **outdated** — the code it pointed at changed.
   That usually means it was addressed, but confirm rather than assume.

3. **If the gate is BLOCKED**, stop. Do not merge. Fix, push, and re-run the gate against the
   new SHA. A fix-up push invalidates the previous review, so say so when reporting.

4. **Merge** — squash is this repo's convention (`develop` history is one commit per PR):

   ```sh
   gh pr merge <pr> --squash --delete-branch
   ```

5. **Verify it landed and sync**:

   ```sh
   gh pr view <pr> --json state,mergedAt
   git checkout develop && git pull --ff-only
   ```

   If the PR closed an issue, confirm the issue actually closed — `Closes #N` only fires when
   the PR merges into the default branch.

## Do not

- **Do not merge on a green tick alone.** That is the failure this skill exists to prevent.
- **Do not lower the check-count floor** to make the gate pass. An empty check array satisfies
  every "all green" test ever written; the floor is what distinguishes "all passed" from "none
  ran".
- **Do not rebase-force-push** to tidy history without asking. Squash-merge already collapses
  fix-up commits into one, so the usual reason to force-push does not apply here.
