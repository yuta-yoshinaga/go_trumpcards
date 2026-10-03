#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# sweep.sh — for each open batch PR: land if reviewed + green; report failures and unreviewed-with-review-posted
require_branch_re || exit $?
branch_re="$BATCH_BRANCH_RE"
for pr in $(gh pr list --state open --author @me --limit 200 --json number,headRefName | jq -r --arg re "$branch_re" '.[]|select(.headRefName|test($re))|.number'); do
  mg=$(gh pr view $pr --json mergeable -q .mergeable); [ "$mg" = CONFLICTING ] && { echo "CONFLICT $pr"; continue; }
  j=$(gh pr checks $pr --json bucket,name 2>/dev/null)
  pend=$(jq '[.[]|select(.bucket=="pending")]|length' <<<"$j"); fail=$(jq -r '[.[]|select(.bucket=="fail")|.name]|join(",")' <<<"$j")
  if [ -n "$fail" ]; then echo "FAIL $pr: $fail"; continue; fi
  if ! grep -qx $pr "$BATCH_STATE/reviewed.txt"; then
    n=$(gh pr view $pr --json comments -q '[.comments[]|select(.author.login=="github-actions")]|length')
    [ "$n" -gt 0 ] && echo "REVIEW-READY $pr"; continue
  fi
  [ "$pend" = 0 ] && "$B/land.sh" "$pr"
done
