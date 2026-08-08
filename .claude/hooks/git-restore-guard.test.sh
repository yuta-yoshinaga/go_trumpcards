#!/usr/bin/env bash
# Self-test for git-restore-guard.sh, run against a throwaway repo in a temp dir.
#
# Every case comes in a pair: one input that must be blocked and one that must
# not. A guard tested only on the failing input passes just as happily when it
# blocks everything, and this one blocks a command people legitimately use, so
# the "must not fire" half is the half that keeps it installed.
set -uo pipefail

HOOK="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/git-restore-guard.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

cd "$TMP" || exit 1
git init -q .
git config user.email t@example.com
git config user.name test

printf 'clean\n' > clean.txt
printf 'dirty\n' > dirty.txt
printf 'staged\n' > staged.txt
mkdir -p docs
printf 'nested\n' > docs/nested.txt
git add -A >/dev/null
git commit -qm init

# dirty.txt has an unstaged edit; staged.txt has a staged one; docs/nested.txt dirty.
printf 'dirty edited\n' > dirty.txt
printf 'staged edited\n' > staged.txt && git add staged.txt
printf 'nested edited\n' > docs/nested.txt

fail=0
# run <expect: block|pass> <description> <command>
run() {
  local expect=$1 desc=$2 cmd=$3 out
  out=$(printf '{"tool_input":{"command":%s}}' "$(printf '%s' "$cmd" | jq -Rs .)" | bash "$HOOK")
  local blocked=pass
  printf '%s' "$out" | grep -q '"continue": *false' && blocked=block
  if [ "$blocked" = "$expect" ]; then
    printf '  ok    %s (%s)\n' "$desc" "$expect"
  else
    printf '  FAIL  %s: expected %s, got %s -- %s\n' "$desc" "$expect" "$blocked" "$out"
    fail=1
  fi
}

echo 'git-restore-guard:'

# --- not our business ----------------------------------------------------
run pass 'unrelated command'                'git status'
run pass 'branch-switching checkout'        'git checkout -b feature/x'
run pass 'plain checkout of a branch'       'git checkout develop'

# --- the destructive forms, on files that would actually lose work --------
run block 'checkout -- <dirty file>'        'git checkout -- dirty.txt'
run block 'checkout -- <dirty nested file>' 'git checkout -- docs/nested.txt'
run block 'checkout <ref> -- <dirty file>'  'git checkout HEAD -- dirty.txt'
run block 'restore <dirty file>'            'git restore dirty.txt'
run block 'checkout -- . with dirt present' 'git checkout -- .'
run block 'restore --staged --worktree'     'git restore --staged --worktree dirty.txt'
run block 'a dirty path among clean ones'   'git checkout -- clean.txt dirty.txt'

# A staged-only change is still overwritten by `git checkout --`, so it counts.
run block 'checkout -- <staged-only file>'  'git checkout -- staged.txt'

# --- the same commands where nothing would be lost -----------------------
run pass 'checkout -- <clean file>'         'git checkout -- clean.txt'
run pass 'restore <clean file>'             'git restore clean.txt'
run pass 'restore --staged only (index)'    'git restore --staged dirty.txt'
run pass 'checkout -- <untracked path>'     'git checkout -- nosuchfile.txt'

if [ "$fail" -ne 0 ]; then
  echo 'git-restore-guard: FAILED'
  exit 1
fi
echo 'git-restore-guard: all cases passed'
