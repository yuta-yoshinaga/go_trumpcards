#!/usr/bin/env bash
# Report coverage for what THIS branch changed, before pushing.
#
# Codecov enforces 80% on project and patch, but only after a push -- and the self-review
# checklist asks for the same number before the work is called done, which today means running
# it by hand or not at all. This runs the two suites scoped to the changed files only, so the
# answer arrives in seconds rather than in a full-suite run.
#
# Go reports STATEMENT coverage (the toolchain has no branch mode); that is the same number
# Codecov gates on, so it is the right proxy, but it is not C1 and this script does not claim
# it is. Vitest's v8 provider does report branch coverage, and that column is what is shown
# for the frontend.
#
# Usage: coverage-gate.sh [base-ref]      (default base: develop)
set -uo pipefail

cd "$(dirname "$0")/../../../.."
BASE="${1:-develop}"
THRESHOLD=80

# Untracked files must be in this list. `git diff` never reports them, and a brand-new source
# file is the single most likely thing on a branch to have no test at all -- the case the gate
# exists for would have been the one case it could not see.
changed=$(git diff --name-only "$BASE...HEAD"; git diff --name-only; git diff --cached --name-only
          git ls-files --others --exclude-standard)
changed=$(printf '%s\n' "$changed" | sort -u | grep -v '^$')

if [ -z "$changed" ]; then
  echo "no changes against $BASE -- nothing to measure."
  exit 0
fi

fail=0

# --- Go -------------------------------------------------------------------------------
# cmd/ and internal/infrastructure/ are excluded from the coverage standard (codecov.yml
# ignores them), so measuring them here would report a number nobody gates on.
pkgs=$(printf '%s\n' "$changed" \
  | grep -E '^(internal|api)/.*\.go$' \
  | grep -v '_test\.go$' \
  | grep -v '^internal/infrastructure/' \
  | xargs -r -n1 dirname | sort -u)

echo "== Go =="
if [ -z "$pkgs" ]; then
  echo "  (no in-scope Go packages changed)"
else
  for p in $pkgs; do
    line=$(go test -tags test -cover "./$p" 2>&1 | tail -1)
    pct=$(printf '%s' "$line" | sed -nE 's/.*coverage: ([0-9.]+)% of statements.*/\1/p')
    if [ -z "$pct" ]; then
      printf '  %-58s %s\n' "$p" "NO RESULT -- $line"
      fail=1
      continue
    fi
    verdict=OK
    awk -v a="$pct" -v b="$THRESHOLD" 'BEGIN{exit !(a+0 < b+0)}' && { verdict="UNDER ${THRESHOLD}%"; fail=1; }
    printf '  %-58s %6s%%  %s\n' "$p" "$pct" "$verdict"
  done
fi

# --- Frontend -------------------------------------------------------------------------
# Per changed source file, run only its own test file and measure only that file. Running the
# whole suite for coverage takes over ten minutes locally; this keeps the answer scoped to the
# question that was asked.
srcs=$(printf '%s\n' "$changed" \
  | grep -E '^frontend/src/(api|components|pages|utils)/.*\.tsx?$' \
  | grep -v '\.test\.tsx\?$')

echo
echo "== Frontend (branch coverage) =="
if [ -z "$srcs" ]; then
  echo "  (no in-scope frontend sources changed)"
else
  for s in $srcs; do
    rel=${s#frontend/}
    test_file="${rel%.*}.test.${rel##*.}"
    if [ ! -f "frontend/$test_file" ]; then
      printf '  %-58s %s\n' "$rel" "NO TEST FILE"
      fail=1
      continue
    fi
    out=$(cd frontend && bunx vitest run "$test_file" --coverage --coverage.reporter=text \
            --coverage.include="$rel" 2>&1)
    # Read the "Coverage summary" block, not the per-file table: with --coverage.include
    # narrowed to a single file the table is emitted with headers and no rows, so a row-based
    # parser reports NO RESULT for a run that measured perfectly well. The summary block is
    # scoped to the same include, so it is the file's own number.
    br=$(printf '%s\n' "$out" | sed -nE 's/^Branches[[:space:]]*:[[:space:]]*([0-9.]+)%.*/\1/p' | head -1)
    if [ -z "$br" ]; then
      printf '  %-58s %s\n' "$rel" "NO RESULT (see: bunx vitest run $test_file --coverage)"
      fail=1
      continue
    fi
    verdict=OK
    awk -v a="$br" -v b="$THRESHOLD" 'BEGIN{exit !(a+0 < b+0)}' && { verdict="UNDER ${THRESHOLD}%"; fail=1; }
    printf '  %-58s %6s%%  %s\n' "$rel" "$br" "$verdict"
  done
fi

echo
if [ "$fail" -eq 0 ]; then
  echo "coverage-gate: PASS (every changed in-scope file at or above ${THRESHOLD}%)"
else
  echo "coverage-gate: FAIL -- add tests for the files marked above before pushing."
fi
exit "$fail"
