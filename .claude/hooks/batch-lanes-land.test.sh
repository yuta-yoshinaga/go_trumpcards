#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPTS="$ROOT/.claude/skills/batch-lanes/scripts"
fail() { echo "FAIL: $*" >&2; exit 1; }
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
REAL_GIT="$(command -v git)"
mkdir -p "$tmp/bin" "$tmp/repo" "$tmp/state"
cat >"$tmp/bin/git" <<'STUB'
#!/usr/bin/env bash
case "$*" in
  *"rev-parse --show-toplevel"*) echo "$BATCH_REPO" ;;
  *"rev-parse --git-common-dir"*) echo .git ;;
  *"rev-parse origin/"*) echo test-sha ;;
  *) exit 0 ;;
esac
STUB
cat >"$tmp/bin/gh" <<'STUB'
#!/usr/bin/env bash
case "$*" in
  "pr view 7 --json headRefName -q .headRefName") echo 'feat/7-test' ;;
  "pr view 7 --json files -q .files[].path") printf '%s\n' "${GH_FILES:-README.md}" ;;
  "pr view 7 --json state -q .state") echo "$GH_STATE" ;;
  "pr view 7 --json closingIssuesReferences -q .closingIssuesReferences[].number") : ;;
  "pr checks 7 --json name,bucket") echo "${GH_CHECKS:-[]}" ;;
  "pr merge 7 --squash") echo called >> "$GH_MERGE_LOG"; exit "${GH_MERGE_RC:-0}" ;;
  "pr list"*) echo '[{"number":8,"headRefName":"fix/10523-old"}]' ;;
  "pr view 8 --json mergeable -q .mergeable") echo MERGEABLE ;;
  "pr checks 8 --json bucket,name") echo '[{"bucket":"pending","name":"CI"}]' ;;
  *) echo "unexpected gh invocation: $*" >&2; exit 2 ;;
esac
STUB
cat >"$tmp/bin/merge-gate" <<'STUB'
#!/usr/bin/env bash
echo 'VERDICT: gate passed'
STUB
chmod +x "$tmp/bin/"*
export PATH="$tmp/bin:$PATH" BATCH_REPO="$tmp/repo" BATCH_STATE="$tmp/state"
export MERGE_GATE="$tmp/bin/merge-gate"
printf '7\n' > "$tmp/state/reviewed.txt"
export GH_MERGE_LOG="$tmp/merge-calls"
GH_MERGE_RC=1 GH_STATE=MERGED bash "$SCRIPTS/land.sh" 7 >"$tmp/merged.out" 2>&1 || fail "land.sh rejected a merge reported as MERGED"
grep -q '^PR 7 MERGED$' "$tmp/merged.out" || fail "land.sh did not report MERGED: $(cat "$tmp/merged.out")"
if GH_MERGE_RC=1 GH_STATE=OPEN bash "$SCRIPTS/land.sh" 7 >"$tmp/open.out" 2>&1; then
  fail "land.sh accepted an OPEN PR after merge failure"
fi
grep -q 'Failed to merge PR 7' "$tmp/open.out" || fail "land.sh did not report the OPEN merge failure"
rm -f "$GH_MERGE_LOG"
: > "$tmp/state/reviewed.txt"
if GH_STATE=OPEN bash "$SCRIPTS/land.sh" 7 >"$tmp/unreviewed.out" 2>&1; then fail "land.sh accepted an unreviewed PR"; fi
grep -q 'REFUSED: PR 7 is not in reviewed.txt' "$tmp/unreviewed.out" || fail "land.sh did not explain unreviewed PR"
[ ! -e "$GH_MERGE_LOG" ] || fail "land.sh merged an unreviewed PR"
printf '7\n' > "$tmp/state/reviewed.txt"
if GH_FILES='internal/foo.go' GH_CHECKS='[]' GH_STATE=OPEN bash "$SCRIPTS/land.sh" 7 >"$tmp/codecov-wait.out" 2>&1; then fail "land.sh accepted missing codecov/patch"; else rc=$?; fi
[ "${rc:-0}" = 3 ] || fail "missing codecov/patch did not return 3"
grep -q 'WAIT: codecov/patch has not reported on PR 7' "$tmp/codecov-wait.out" || fail "land.sh did not report missing codecov"
[ ! -e "$GH_MERGE_LOG" ] || fail "land.sh merged before codecov/patch reported"
if GH_FILES='internal/foo.go' GH_CHECKS='[{"name":"codecov/patch","bucket":"pending"}]' GH_STATE=OPEN bash "$SCRIPTS/land.sh" 7 >"$tmp/codecov-pending.out" 2>&1; then fail "land.sh accepted pending codecov/patch"; else rc=$?; fi
[ "${rc:-0}" = 3 ] || fail "pending codecov/patch did not return 3"
grep -q 'WAIT: codecov/patch is still pending on PR 7' "$tmp/codecov-pending.out" || fail "land.sh did not report pending codecov"
[ ! -e "$GH_MERGE_LOG" ] || fail "land.sh merged while codecov/patch was pending"
if GH_FILES='internal/foo.go' GH_CHECKS='[{"name":"codecov/patch","bucket":"fail"}]' GH_STATE=OPEN bash "$SCRIPTS/land.sh" 7 >"$tmp/codecov-fail.out" 2>&1; then fail "land.sh accepted failing codecov/patch"; fi
grep -q 'REFUSED: codecov/patch is fail' "$tmp/codecov-fail.out" || fail "land.sh did not report failing codecov"
[ ! -e "$GH_MERGE_LOG" ] || fail "land.sh merged with failing codecov/patch"
GH_FILES='internal/foo.go' GH_CHECKS='[{"name":"codecov/patch","bucket":"pass"}]' GH_STATE=MERGED \
  bash "$SCRIPTS/land.sh" 7 >"$tmp/codecov-pass.out" 2>&1 || fail "land.sh rejected passing codecov/patch"
[ -e "$GH_MERGE_LOG" ] || fail "land.sh did not merge with passing codecov/patch"
rm -f "$GH_MERGE_LOG"
GH_FILES='README.md' GH_CHECKS='[]' GH_STATE=MERGED bash "$SCRIPTS/land.sh" 7 >"$tmp/docs.out" 2>&1 || fail "land.sh rejected a docs-only PR without codecov"
[ -e "$GH_MERGE_LOG" ] || fail "land.sh did not merge docs-only PR"
BATCH_BRANCH_RE='^feat/(10[1-5][0-9]{2})-' bash "$SCRIPTS/sweep.sh" >"$tmp/sweep.out" || fail "sweep.sh failed"
grep -q '^UNSWEPT 8 fix/10523-old$' "$tmp/sweep.out" || fail "sweep.sh did not report the unmatched PR: $(cat "$tmp/sweep.out")"
BATCH_BRANCH_RE='^feat/(10[1-5][0-9]{2})-' BATCH_UNSWEPT_IGNORE='8 999' bash "$SCRIPTS/sweep.sh" >"$tmp/sweep-ignored.out" || fail "sweep.sh failed with BATCH_UNSWEPT_IGNORE"
! grep -q '^UNSWEPT 8' "$tmp/sweep-ignored.out" || fail "sweep.sh reported ignored PR 8: $(cat "$tmp/sweep-ignored.out")"

# land rejects dirty and unpushed PR worktrees, with an explicit pending override.
(
  git() { "$REAL_GIT" "$@"; }
  repo="$tmp/real-land-repo"; origin="$tmp/real-land-origin.git"; wt="$tmp/real-land-wts/wt-ib7"
  mkdir -p "$repo" "$(dirname "$wt")" "$tmp/real-land-state" "$tmp/real-land-bin"
  git init -q "$repo"; git -C "$repo" config user.email test@example.com; git -C "$repo" config user.name test
  printf 'base\n' > "$repo/file"; git -C "$repo" add file; git -C "$repo" commit -qm base; git -C "$repo" branch -M develop
  git init -q --bare "$origin"; git -C "$origin" symbolic-ref HEAD refs/heads/develop
  git -C "$repo" remote add origin "$origin"; git -C "$repo" push -qu origin develop
  git -C "$repo" fetch -q origin
  git -C "$repo" worktree add -qb feat/7-test "$wt" origin/develop
  [ -f "$wt/file" ] || fail "land fixture worktree was not created: $(git -C "$repo" worktree list)"
  git -C "$wt" push -qu origin HEAD:feat/7-test
  git -C "$repo" fetch -q origin feat/7-test:refs/remotes/origin/feat/7-test
  printf '7\n' > "$tmp/real-land-state/reviewed.txt"
  cat > "$tmp/real-land-bin/gh" <<'GH'
#!/usr/bin/env bash
case "$*" in
  'pr view 7 --json files -q .files[].path') echo README.md ;;
  'pr view 7 --json headRefName -q .headRefName') echo feat/7-test ;;
  'pr merge 7 --squash') echo called >> "$GH_MERGE_LOG" ;;
  'pr view 7 --json state -q .state') echo MERGED ;;
  'pr view 7 --json closingIssuesReferences -q .closingIssuesReferences[].number') : ;;
  *) echo "unexpected gh call: $*" >&2; exit 2 ;;
esac
GH
  cat > "$tmp/real-land-bin/merge-gate" <<'GATE'
#!/usr/bin/env bash
echo 'VERDICT: gate passed'
GATE
  cat > "$tmp/real-land-bin/git" <<'GIT'
#!/usr/bin/env bash
exec "$REAL_GIT" "$@"
GIT
  chmod +x "$tmp/real-land-bin/"*
  export REAL_GIT
  export GH_MERGE_LOG="$tmp/real-land-merges"
  run_land() { PATH="$tmp/real-land-bin:$PATH" BATCH_REPO="$repo" BATCH_STATE="$tmp/real-land-state" BATCH_WT_ROOT="$(dirname "$wt")" MERGE_GATE="$tmp/real-land-bin/merge-gate" bash "$SCRIPTS/land.sh" 7; }
  BUSY_ARGV0="codex-task -C $wt " bash -c 'exec -a "$BUSY_ARGV0" sleep 30' &
  busy_pid=$!
  sleep 0.1
  if run_land >"$tmp/land-busy.out" 2>&1; then kill "$busy_pid" 2>/dev/null || true; fail 'land accepted a busy delegation worktree'; fi
  kill "$busy_pid" 2>/dev/null || true
  wait "$busy_pid" 2>/dev/null || true
  grep -q 'delegation is still running' "$tmp/land-busy.out" || fail "land busy refusal reason missing: $(cat "$tmp/land-busy.out")"
  [ ! -e "$GH_MERGE_LOG" ] || fail 'land merged while delegation was running'
  printf 'dirty\n' >> "$wt/file"
  if run_land >"$tmp/land-dirty.out" 2>&1; then fail 'land accepted uncommitted worktree changes'; fi
  grep -q 'uncommitted changes' "$tmp/land-dirty.out" || fail "land dirty refusal reason missing: $(cat "$tmp/land-dirty.out")"
  [ ! -e "$GH_MERGE_LOG" ] || fail 'land merged despite uncommitted changes'
  git -C "$wt" show HEAD:file > "$wt/file"
  printf 'committed\n' >> "$wt/file"; git -C "$wt" commit -qam pending
  if run_land >"$tmp/land-unpushed.out" 2>&1; then fail 'land accepted an unpushed commit'; fi
  grep -q 'unpushed commit' "$tmp/land-unpushed.out" || fail "land unpushed refusal reason missing: $(cat "$tmp/land-unpushed.out")"
  [ ! -e "$GH_MERGE_LOG" ] || fail 'land merged despite an unpushed commit'
  git -C "$wt" push -qu origin HEAD:feat/7-test
  git -C "$repo" fetch -q origin feat/7-test:refs/remotes/origin/feat/7-test
  run_land >"$tmp/land-clean.out" 2>&1 || fail "land rejected clean pushed worktree: $(cat "$tmp/land-clean.out")"
  [ -e "$GH_MERGE_LOG" ] || fail 'land did not merge clean pushed worktree'
  rm -f "$GH_MERGE_LOG"
  git -C "$origin" update-ref -d refs/heads/feat/7-test
  if run_land >"$tmp/land-missing-origin.out" 2>&1; then fail 'land accepted a worktree whose origin branch is missing'; fi
  grep -q "REFUSED: $wt cannot compare with origin/feat/7-test" "$tmp/land-missing-origin.out" || fail "land did not explain missing origin branch: $(cat "$tmp/land-missing-origin.out")"
  git -C "$wt" push -qu origin HEAD:feat/7-test
  git -C "$repo" fetch -q origin feat/7-test:refs/remotes/origin/feat/7-test
  printf 'dirty again\n' >> "$wt/file"
  ALLOW_LANE_PENDING=1 run_land >"$tmp/land-override.out" 2>&1 || fail "land rejected ALLOW_LANE_PENDING: $(cat "$tmp/land-override.out")"
  [ -e "$GH_MERGE_LOG" ] || fail 'land override did not merge'
  rm -f "$GH_MERGE_LOG"
  git -C "$repo" worktree remove --force "$wt"
  run_land >"$tmp/land-no-worktree.out" 2>&1 || fail "land rejected absent PR worktree: $(cat "$tmp/land-no-worktree.out")"
  [ -e "$GH_MERGE_LOG" ] || fail 'land did not merge when PR worktree was absent'
) || exit 1
echo "batch-lanes land tests passed"
