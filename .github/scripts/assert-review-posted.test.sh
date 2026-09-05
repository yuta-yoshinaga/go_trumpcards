#!/usr/bin/env bash
# Unit tests for assert-review-posted.sh
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET="$SCRIPT_DIR/assert-review-posted.sh"

test_count=0
fail_count=0

run_test() {
  local expect="$1" desc="$2" mode="$3" input="$4" url="${5:-https://github.com/example/repo/pull/1#issuecomment-123}"
  test_count=$((test_count + 1))

  local actual="pass"
  local stdout_file stderr_file
  stdout_file=$(mktemp)
  stderr_file=$(mktemp)

  if [ "$mode" = "stdin" ]; then
    if ! printf '%s\n' "$input" | bash "$TARGET" --stdin "$url" >"$stdout_file" 2>"$stderr_file"; then
      actual="fail"
    fi
  elif [ "$mode" = "file" ]; then
    local tmp_input
    tmp_input=$(mktemp)
    printf '%s\n' "$input" > "$tmp_input"
    if ! bash "$TARGET" --file "$tmp_input" "$url" >"$stdout_file" 2>"$stderr_file"; then
      actual="fail"
    fi
    rm -f "$tmp_input"
  elif [ "$mode" = "missing-file" ]; then
    if ! bash "$TARGET" --file "/path/to/nonexistent/file/for/test" "$url" >"$stdout_file" 2>"$stderr_file"; then
      actual="fail"
    fi
  fi

  local err_out
  err_out=$(cat "$stderr_file")
  rm -f "$stdout_file" "$stderr_file"

  if [ "$actual" = "$expect" ]; then
    echo "  ok    $desc ($expect)"
  else
    echo "  FAIL  $desc: expected $expect, got $actual" >&2
    [ -n "$err_out" ] && echo "        stderr: $err_out" >&2
    fail_count=$((fail_count + 1))
  fi
}

run_test_with_stderr_check() {
  local desc="$1" input="$2" url="$3" expected_incomplete="$4"
  test_count=$((test_count + 1))

  local stderr_file
  stderr_file=$(mktemp)

  if printf '%s\n' "$input" | bash "$TARGET" --stdin "$url" >/dev/null 2>"$stderr_file"; then
    echo "  FAIL  $desc: expected exit non-zero, got 0" >&2
    rm -f "$stderr_file"
    fail_count=$((fail_count + 1))
    return
  fi

  local err_out
  err_out=$(cat "$stderr_file")
  rm -f "$stderr_file"

  if ! printf '%s\n' "$err_out" | grep -Fq -- "$url"; then
    echo "  FAIL  $desc: stderr missing URL: $url" >&2
    echo "        stderr: $err_out" >&2
    fail_count=$((fail_count + 1))
    return
  fi

  if ! printf '%s\n' "$err_out" | grep -Fq -- "$expected_incomplete"; then
    echo "  FAIL  $desc: stderr missing expected incomplete item: $expected_incomplete" >&2
    echo "        stderr: $err_out" >&2
    fail_count=$((fail_count + 1))
    return
  fi

  echo "  ok    $desc (fail + stderr verified)"
}

echo "assert-review-posted tests:"

# --- Case 1: Completed review comment (- [x] only) -> pass ---
run_test pass "single completed checklist item via stdin" stdin "- [x] Initial review done"
run_test pass "multiple completed checklist items with summary via stdin" stdin \
"### Review
- [x] Gather context
- [x] Review domain logic
- [x] Post summary

All tests pass and code looks good."
run_test pass "completed checklist items via file" file \
"### Review
- [x] Step 1
- [x] Step 2"

# --- Case 2: Incomplete review comment (- [ ] present) -> fail ---
run_test fail "single unchecked item via stdin" stdin "- [ ] Incomplete task"
run_test fail "mixed completed and incomplete items (PR #7111 pattern)" stdin \
"### Review
- [x] Step 1
- [ ] Step 2: unfinished work
- [x] Step 3"
run_test fail "indented incomplete item" stdin \
"### Review
  - [ ] Nested incomplete task"
run_test fail "asterisk bullet incomplete item" stdin \
"### Review
* [ ] Asterisk bullet task"
run_test fail "incomplete item via file" file \
"### Review
- [ ] File-based incomplete task"

# Verify stderr contains URL and incomplete item text
run_test_with_stderr_check "stderr includes comment URL and incomplete item content" \
"### Reviewing PR #7111
- [x] Gather context
- [ ] Review i18n parity, registration completeness
- [ ] Post final grouped review summary" \
"https://github.com/yuta-yoshinaga/go_trumpcards/pull/7111#issuecomment-5550850494" \
"- [ ] Review i18n parity, registration completeness"

# --- Case 3: Empty / missing comment -> fail ---
run_test fail "empty body via stdin" stdin ""
run_test fail "whitespace-only body via stdin" stdin "   
   	
   "
run_test fail "empty file via file" file ""
run_test fail "missing file via file" missing-file ""

# --- Guard: Zero tests executed must fail ---
if [ "$test_count" -eq 0 ]; then
  echo "FAIL: 0 tests were executed" >&2
  exit 1
fi

if [ "$fail_count" -gt 0 ]; then
  echo "FAILED: $fail_count of $test_count test(s) failed" >&2
  exit 1
fi

echo "All $test_count tests passed."
exit 0
