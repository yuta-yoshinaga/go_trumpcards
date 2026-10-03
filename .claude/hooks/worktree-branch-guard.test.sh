#!/usr/bin/env bash
set -uo pipefail
HOOK="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/worktree-branch-guard.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
git init -q "$TMP/main"
cd "$TMP/main" || exit 1
git config user.email t@example.com
git config user.name test
# Build a minimal commit object without invoking a repository commit command.
TREE=$(git mktree </dev/null)
COMMIT=$(printf 'tree %s\nauthor test <t@example.com> 0 +0000\ncommitter test <t@example.com> 0 +0000\n\ninit\n' "$TREE" | git hash-object -t commit -w --stdin)
git update-ref refs/heads/main "$COMMIT"
git symbolic-ref HEAD refs/heads/main
git branch shared
git worktree add -q "$TMP/other" shared
fail=0
run() {
  local expect=$1 desc=$2 cmd=$3 out blocked=pass
  out=$(printf '{"tool_input":{"command":%s}}' "$(printf '%s' "$cmd" | jq -Rs .)" | bash "$HOOK")
  printf '%s' "$out" | grep -q '"permissionDecision": *"deny"' && blocked=block
  if [ "$blocked" = "$expect" ]; then echo "  ok $desc ($expect)"; else echo "  FAIL $desc expected=$expect got=$blocked $out"; fail=1; fi
  if [ "$blocked" = block ]; then
    printf '%s' "$out" | grep -Fq "$TMP/other" || { echo "  FAIL $desc missing owning worktree"; fail=1; }
    printf '%s' "$out" | grep -q -- '--detach\|that worktree\|worktree' || { echo "  FAIL $desc missing remedy"; fail=1; }
    ! printf '%s' "$out" | grep -q '"continue" *: *false' || { echo "  FAIL $desc used continue false"; fail=1; }
  fi
}
echo 'worktree-branch-guard:'
run block 'checkout force other worktree branch' 'git checkout -B shared origin/shared'
run block 'switch force other worktree branch' 'git switch -C shared'
run block 'branch force other worktree branch' 'git branch -f shared origin/shared'
run block 'branch long force' 'git branch --force shared'
run block 'update-ref other worktree branch' 'git update-ref refs/heads/shared HEAD'
run block 'fetch refspec other worktree branch' 'git fetch origin feature:shared'
run block 'git -C current worktree syntax' "git -C '$TMP/main' checkout -B shared"
run pass 'unoccupied branch' 'git checkout -B free origin/free'
run pass 'current worktree branch' 'git checkout -B main'
run pass 'detach checkout' 'git checkout shared --detach'
run pass 'fetch without destination' 'git fetch origin develop'
[ "$fail" -eq 0 ] || exit 1
echo 'worktree-branch-guard: all cases passed'
