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
  elif [ "$mode" = "missing-comments-file" ]; then
    if ! bash "$TARGET" --comments-file "/path/to/nonexistent/file/for/test" >"$stdout_file" 2>"$stderr_file"; then
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

run_selection_test() {
  local expect="$1" desc="$2" json="$3" expected_match="${4:-}"
  test_count=$((test_count + 1))

  local actual="pass"
  local stdout_file stderr_file tmp_json
  stdout_file=$(mktemp)
  stderr_file=$(mktemp)
  tmp_json=$(mktemp)

  printf '%s\n' "$json" > "$tmp_json"

  if ! bash "$TARGET" --comments-file "$tmp_json" >"$stdout_file" 2>"$stderr_file"; then
    actual="fail"
  fi
  rm -f "$tmp_json"

  local stdout_out stderr_out
  stdout_out=$(cat "$stdout_file")
  stderr_out=$(cat "$stderr_file")
  rm -f "$stdout_file" "$stderr_file"

  if [ "$actual" != "$expect" ]; then
    echo "  FAIL  $desc: expected $expect, got $actual" >&2
    [ -n "$stderr_out" ] && echo "        stderr: $stderr_out" >&2
    fail_count=$((fail_count + 1))
    return
  fi

  if [ "$expect" = "pass" ] && [ -n "$expected_match" ]; then
    if ! printf '%s\n' "$stdout_out" | grep -Fq -- "$expected_match"; then
      echo "  FAIL  $desc: output missing expected match: $expected_match" >&2
      echo "        stdout: $stdout_out" >&2
      fail_count=$((fail_count + 1))
      return
    fi
  fi

  echo "  ok    $desc ($expect)"
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

# --- Case 4: Comment selection logic via --comments-file ---
# Subcase 4.1: Mixed comments with unrelated bots -> tracking comment selected
run_selection_test pass "tracking comment selected when mixed with other bot comments" \
'[[
  {"id": 1001, "user": {"login": "dependabot[bot]"}, "body": "Bumps actions/checkout from 3 to 4"},
  {"id": 1002, "user": {"login": "github-actions[bot]"}, "body": "**Claude finished @user'\''s task in 2m**\n\n### Review\n- [x] All checks pass"},
  {"id": 1003, "user": {"login": "stale[bot]"}, "body": "This issue is marked stale."}
]]' '"id": 1002'

# Subcase 4.2: Codecov decoy with "review" in body -> must not be selected
run_selection_test fail "codecov decoy comment with review in body is not selected" \
'[[
  {"id": 2001, "user": {"login": "codecov[bot]"}, "body": "## [Codecov](https://example.com) Report\nPatch coverage is 93%. Please review.\n| Files | Coverage |"}
]]'

run_selection_test pass "tracking comment selected over subsequent codecov decoy" \
'[[
  {"id": 2002, "user": {"login": "github-actions[bot]"}, "body": "**Claude finished @user'\''s task**\n\n- [x] Done"},
  {"id": 2003, "user": {"login": "codecov[bot]"}, "body": "## [Codecov](https://example.com) Report\nPatch coverage is 93%. Please review."}
]]' '"id": 2002'

# Subcase 4.3: github-actions[bot] comment without **Claude header -> must not be selected
run_selection_test fail "github-actions[bot] comment without Claude header is not selected" \
'[[
  {"id": 3001, "user": {"login": "github-actions[bot]"}, "body": "Automated workflow review finished. No action needed."}
]]'

run_selection_test pass "tracking comment selected over subsequent github-actions[bot] decoy" \
'[[
  {"id": 3002, "user": {"login": "github-actions[bot]"}, "body": "**Claude finished @user'\''s task**\n\n- [x] Done"},
  {"id": 3003, "user": {"login": "github-actions[bot]"}, "body": "Automated workflow review finished. All steps passed."}
]]' '"id": 3002'

# Subcase 4.4: No tracking comments present -> fail
run_selection_test fail "empty slurped pages array produces error" \
'[[]]'

run_selection_test fail "only unrelated human comments produces error" \
'[[
  {"id": 4001, "user": {"login": "alice"}, "body": "LGTM, nice work!"}
]]'

# Subcase 4.5: Multiple tracking comments -> last one is selected
run_selection_test pass "last tracking comment selected when multiple tracking comments exist" \
'[[
  {"id": 5001, "user": {"login": "github-actions[bot]"}, "body": "**Claude finished first attempt**\n- [ ] Fix pending"},
  {"id": 5002, "user": {"login": "github-actions[bot]"}, "body": "**Claude finished final review**\n- [x] Completed"}
]]' '"id": 5002'

# Subcase 4.6: Multi-page comments -> last tracking comment across pages selected
run_selection_test pass "last tracking comment selected across multiple slurped pages" \
'[[
  {"id": 6001, "user": {"login": "github-actions[bot]"}, "body": "**Claude finished page 1**\n- [ ] WIP"}
], [
  {"id": 6002, "user": {"login": "claude[bot]"}, "body": "**Claude finished page 2**\n- [x] Final review done"}
]]' '"id": 6002'

# Subcase 4.7: claude[bot] author is accepted
run_selection_test pass "tracking comment posted by claude[bot] is selected" \
'[[
  {"id": 7001, "user": {"login": "claude[bot]"}, "body": "**Claude finished task**\n- [x] Complete"}
]]' '"id": 7001'

# Subcase 4.8: Missing file via --comments-file -> fail
run_test fail "missing file via --comments-file" missing-comments-file ""

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
