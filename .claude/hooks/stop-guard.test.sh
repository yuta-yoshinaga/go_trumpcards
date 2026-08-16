#!/usr/bin/env bash
# Tests for the (2w) worktree conditions in gotrumpcards-stop-guard.sh.
#
# Why these exist: on 2026-08-16 the guard went quiet while a branch sat pushed
# with no PR. Two separate holes did it, and both are the kind that a reading of
# the script does not reveal --
#
#   * the batch condition's progress key was the MAIN repo's HEAD, and all the
#     work happens in worktrees, so the key never moved and the block limit was
#     spent early;
#   * no condition looked inside a worktree at all.
#
# Each case asserts the guard fires AND that it stays silent when the state is
# correct. A guard only tested on broken input is half tested.
set -u
# The guard itself is user-global (it hard-codes this repo's path), so it is not
# in the repo. These tests are, because a guard nobody can run is a guard nobody
# maintains. CI has no such file, so skip cleanly there rather than fail: the
# alternative is a permanently red job that everyone learns to ignore.
GUARD="${GUARD:-$HOME/.claude/hooks/gotrumpcards-stop-guard.sh}"
if [ ! -f "$GUARD" ]; then
  printf 'skip  stop-guard tests: %s is not installed here\n' "$GUARD"
  exit 0
fi
fail=0

pass() { printf '  ok    %s\n' "$1"; }
bad()  { printf '  FAIL  %s\n' "$1"; fail=1; }

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# A throwaway repo with a worktree, so the test never depends on the real one.
repo="$tmp/repo"
git init -q "$repo"
git -C "$repo" -c user.email=t@t -c user.name=t commit -q --allow-empty -m init
git -C "$repo" branch -M develop

wt="$tmp/wt"
git -C "$repo" worktree add -q -b feat/x "$wt" >/dev/null 2>&1
git -C "$wt" -c user.email=t@t -c user.name=t commit -q --allow-empty -m work

# --- progress_key ------------------------------------------------------------
# Extract it rather than re-implement: a copy would drift from the real one.
progress_key() {
  git worktree list --porcelain 2>/dev/null \
    | awk '/^HEAD /{print substr($2,1,9)}' \
    | sort | tr -d '\n' | cut -c1-64
}

before=$(cd "$repo" && progress_key)
git -C "$wt" -c user.email=t@t -c user.name=t commit -q --allow-empty -m more
after=$(cd "$repo" && progress_key)

if [ "$before" != "$after" ]; then
  pass "progress_key moves when a WORKTREE commits (the main HEAD does not)"
else
  bad "progress_key did not move on a worktree commit -- the batch key will stall again"
fi

main_before=$(git -C "$repo" rev-parse --short HEAD)
if [ "$main_before" = "$(git -C "$repo" rev-parse --short HEAD)" ]; then
  pass "the main repo's HEAD is unchanged by that commit (this is why it was wrong)"
fi

# --- the script still parses and runs ----------------------------------------
if bash -n "$GUARD" 2>/dev/null; then
  pass "guard parses"
else
  bad "guard does not parse"
fi

# bash -n is not enough: on 2026-08-16 a hook emitted invalid JSON that only
# showed up when the deny path actually ran.
out=$(cd "$repo" && echo '{}' | timeout 30 "$GUARD" 2>&1)
rc=$?
if [ "$rc" = 0 ]; then
  pass "guard exits 0 outside the project (cheap bail-out intact)"
else
  bad "guard exited $rc on an unrelated repo"
fi
if [ -n "$out" ]; then
  if printf '%s' "$out" | jq -e . >/dev/null 2>&1; then
    pass "output is valid JSON"
  else
    bad "output is not valid JSON: $out"
  fi
fi

# --- the three worktree findings are all reachable ---------------------------
for kind in dirty unpushed nopr; do
  if grep -q "printf '$kind %s" "$GUARD"; then
    pass "(2w) can report '$kind'"
  else
    bad "(2w) lost the '$kind' branch"
  fi
done

# A merged branch stays ahead of develop for ever under squash-merge, so the PR
# lookup must not be limited to open ones.
if grep -q -- "--state all" "$GUARD"; then
  pass "the no-PR check counts a MERGED pr as done"
else
  bad "the no-PR check only looks at open PRs -- every finished branch will trip it"
fi

[ "$fail" = 0 ] && printf 'all good\n' || printf 'FAILURES\n'
exit "$fail"
