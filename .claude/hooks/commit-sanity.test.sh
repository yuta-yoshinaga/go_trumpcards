#!/usr/bin/env bash
# Self-test for commit-sanity.sh, run against a throwaway repo in a temp dir.
#
# Every case comes in a pair: one input that must be blocked and one that must
# not. A guard tested only on the failing input passes just as happily when it
# blocks everything, which would get it switched off within a day.
set -uo pipefail

HOOK="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/commit-sanity.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

cd "$TMP" || exit 1
git init -q .
git config user.email t@example.com
git config user.name test

fail=0
# run <expect: block|pass> <description> <command>
run() {
  local expect=$1 desc=$2 cmd=$3 out
  out=$(printf '{"tool_input":{"command":%s}}' "$(printf '%s' "$cmd" | jq -Rs .)" | bash "$HOOK")
  local blocked=pass
  # A guard denies one tool call; it must never use the `continue:false` kill
  # switch, which ends the whole turn and whose stopReason is never shown to
  # Claude. Assert the shape, not just "something was refused".
  printf '%s' "$out" | grep -q '"permissionDecision": *"deny"' && blocked=block
  if printf '%s' "$out" | grep -q '"continue" *: *false'; then
    printf '  FAIL  %s: hook used the continue:false kill switch\n' "$desc"
    fail=1
  fi
  if [ "$blocked" = "$expect" ]; then
    printf '  ok    %s (%s)\n' "$desc" "$expect"
  else
    printf '  FAIL  %s: expected %s, got %s -- %s\n' "$desc" "$expect" "$blocked" "$out"
    fail=1
  fi
}

echo 'commit-sanity:'

run pass 'non-commit command is ignored' 'git status'

# --- empty index ---------------------------------------------------------
run block 'commit with nothing staged' 'git commit -m "x"'
run pass  'amend with nothing staged' 'git commit --amend --no-edit'
run pass  'explicit --allow-empty' 'git commit --allow-empty -m "x"'

# --- clean staged content (false-positive control) ------------------------
printf 'const a = 1;\n' > clean.ts
cat > underline.md <<'EOF'
Heading
=======

Body text with <<< and >>> in prose.
EOF
git add clean.ts underline.md
run pass 'clean staged content, incl. a Markdown === underline' 'git commit -m "x"'

# --- conflict markers -----------------------------------------------------
# Written with printf rather than a heredoc on purpose. A heredoc puts the marker at column 0
# of THIS file, and this file gets committed like any other -- the guard would then block every
# commit that touches its own test. The fixture on disk is byte-identical either way; only the
# test source differs, and here no line begins with a marker.
printf '<<<<<<< HEAD\nconst a = 1;\n=======\nconst a = 2;\n>>>>>>> feature\n' > conflicted.ts
git add conflicted.ts
run block 'staged conflict markers' 'git commit -m "x"'

exit "$fail"
