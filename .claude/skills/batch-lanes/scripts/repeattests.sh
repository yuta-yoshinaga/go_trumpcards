#!/bin/bash
# repeattests.sh [--list] <worktree> — rerun the Go test functions this change adds or edits,
# many times, so a test that depends on the shuffle fails in the lane instead of on develop.
#
# A single `go test` run passes a deal-dependent test most of the time. Two of them reached
# develop in one evening (2026-10-09): Minibridge played the dealt card at index 0 (#11597),
# French Tarot expected a chien reveal a CPU's Garde Sans can skip (#11602). Both then failed
# unrelated PRs. -count=${REPEAT_TESTS_COUNT:-50} catches the first kind in seconds; a
# 1-in-2000 flake like the second still needs the shuffle-determinism-auditor.
#
# Prints "WARN DEAL_DEPENDENT <pkg> <tests>" plus the first failure lines when a rerun fails.
# --list prints "<pkg dir> <TestName>" per test and runs nothing (used by the hook tests).
set -uo pipefail
list=0
if [[ "${1:-}" == --list ]]; then list=1; shift; fi
wt="$1"
cd "$wt" || exit 1

tests=$(mktemp)
trap 'rm -f "$tests"' EXIT

# Test names per package dir: new `func Test…` lines, plus functions whose body changed
# (git puts the enclosing func in the hunk header), plus every test in untracked files.
# Only internal/ is scanned: the game logic and its deal-dependent tests live there, and
# cmd/ holds entry points. The awk is POSIX (no gawk-only 3-argument match) so it works
# under mawk, Ubuntu's default awk.
{
  git diff -U0 HEAD -- 'internal/*_test.go' | awk '
    /^\+\+\+ \/dev\/null/ { f = ""; next }
    /^\+\+\+ b\// { f = substr($2, 3); next }
    /^@@/ {
      if (f != "" && match($0, /@@ func Test[A-Za-z0-9_]+\(/)) {
        name = substr($0, RSTART + 8, RLENGTH - 9); print f, name
      }
      next
    }
    /^\+func Test[A-Za-z0-9_]+\(/ { if (f != "") { name = $2; sub(/\(.*/, "", name); print f, name } }'
  git ls-files -o --exclude-standard -- 'internal/*_test.go' | while IFS= read -r f; do
    grep -oE '^func Test[A-Za-z0-9_]+\(' "$f" | sed -E 's/^func //; s/\($//' | sed "s|^|$f |"
  done
} | awk '{ d = $1; sub(/\/[^\/]*$/, "", d); print d, $2 }' | sort -u > "$tests"

if ((list)); then cat "$tests"; exit 0; fi

count="${REPEAT_TESTS_COUNT:-50}"
for dir in $(cut -d' ' -f1 "$tests" | sort -u); do
  names=$(awk -v d="$dir" '$1 == d { print $2 }' "$tests" | paste -sd'|')
  [[ -n "$names" ]] || continue
  out=$(go test -tags test "./$dir" -run "^($names)\$" -count="$count" 2>&1)
  if [[ $? -ne 0 ]]; then
    if grep -q -- '--- FAIL' <<<"$out"; then
      echo "WARN DEAL_DEPENDENT $dir $names (failed within -count=$count; the test likely depends on the deal)"
    else
      echo "WARN REPEAT_TESTS_ERROR $dir $names (go test failed without a test failure: build or vet error)"
    fi
    grep -E -- '--- FAIL|Error:|_test\.go:[0-9]+|\.go:[0-9]+:[0-9]+:' <<<"$out" | head -6 | sed 's/^/  /'
  fi
done
