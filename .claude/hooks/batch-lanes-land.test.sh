#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPTS="$ROOT/.claude/skills/batch-lanes/scripts"
fail() { echo "FAIL: $*" >&2; exit 1; }
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
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
echo "batch-lanes land tests passed"
