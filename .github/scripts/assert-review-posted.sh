#!/usr/bin/env bash
# assert-review-posted.sh
# Verifies that a Claude Code review tracking comment was posted on a PR
# and that all checklist items are completed (- [x]), with no incomplete (- [ ]) items left.
set -uo pipefail

usage() {
  cat <<'EOF' >&2
Usage:
  assert-review-posted.sh <PR_NUMBER> <REPO>
  assert-review-posted.sh --stdin [COMMENT_URL]
  assert-review-posted.sh --file <FILE> [COMMENT_URL]
EOF
  exit 2
}

check_comment_body() {
  local body="$1"
  local url="${2:-<unknown>}"

  # Empty or whitespace-only body fails
  local trimmed
  trimmed=$(printf '%s' "$body" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')
  if [ -z "$trimmed" ]; then
    echo "ERROR: Review comment body is empty or missing." >&2
    return 1
  fi

  # Check for incomplete checklist items: lines matching "- [ ]" or "* [ ]"
  local incomplete
  incomplete=$(printf '%s\n' "$body" | grep -E '^[[:space:]]*[-*][[:space:]]+\[ \]' || true)

  if [ -n "$incomplete" ]; then
    echo "ERROR: Review tracking comment has incomplete checklist items." >&2
    echo "Comment URL: $url" >&2
    echo "Incomplete items:" >&2
    printf '%s\n' "$incomplete" >&2
    return 1
  fi

  echo "OK: Review tracking comment verified. All checklist items complete."
  [ -n "$url" ] && echo "URL: $url"
  return 0
}

# Standalone body testing mode: --stdin
if [ "${1:-}" = "--stdin" ]; then
  url="${2:-<stdin>}"
  body=$(cat)
  check_comment_body "$body" "$url"
  exit $?
fi

# Standalone body testing mode: --file
if [ "${1:-}" = "--file" ]; then
  if [ $# -lt 2 ]; then
    usage
  fi
  file="$2"
  url="${3:-$file}"
  if [ ! -f "$file" ]; then
    echo "ERROR: File not found: $file" >&2
    exit 2
  fi
  body=$(cat "$file")
  check_comment_body "$body" "$url"
  exit $?
fi

# Normal PR check mode: <PR_NUMBER> <REPO>
if [ $# -lt 2 ]; then
  usage
fi

PR_NUMBER="$1"
REPO="$2"

if ! command -v gh >/dev/null 2>&1; then
  echo "ERROR: gh CLI is required but not installed." >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq is required but not installed." >&2
  exit 1
fi

# Fetch comments via gh api.
#
# Match the tracking comment by the header the action itself prepends
# ("**Claude finished @user's task**"), not by a loose /Claude|Review/i on the
# body. Every workflow that comments with GITHUB_TOKEN posts as
# github-actions[bot], and unrelated bodies do mention "review" -- codecov's
# report on PR #7111 matches /review/i and would have been picked as the
# newest "review" comment if the author filter had not happened to exclude it.
# Anchoring on the header means an unrelated bot comment can never stand in for
# a review that was never posted.
#
# If the action ever changes that header, this reports "no review comment
# found" and the job goes red. That is the safe direction to fail: the bug this
# guard exists for is a review that silently counts as success.
comment_json=$(gh api "repos/${REPO}/issues/${PR_NUMBER}/comments" --jq '
  [ .[] | select((.user.login == "github-actions[bot]" or .user.login == "claude[bot]") and (.body | test("^\\*\\*Claude "))) ] | last // empty
')

if [ -z "$comment_json" ] || [ "$comment_json" = "null" ]; then
  echo "ERROR: No review tracking comment found for PR #${PR_NUMBER} in ${REPO}." >&2
  exit 1
fi

url=$(printf '%s' "$comment_json" | jq -r '.html_url // empty')
body=$(printf '%s' "$comment_json" | jq -r '.body // empty')

check_comment_body "$body" "$url"
