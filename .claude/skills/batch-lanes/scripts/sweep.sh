#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# sweep.sh — for each open batch PR: land if reviewed + green; report failures and unreviewed-with-review-posted
require_branch_re || exit $?
branch_re="$BATCH_BRANCH_RE"
prs=$(gh pr list --state open --author @me --limit 200 --json number,headRefName)
while IFS=$'\t' read -r pr branch; do
  [[ -z "$pr" ]] && continue
  if ! jq -e --arg re "$branch_re" -n --arg branch "$branch" '$branch|test($re)' >/dev/null; then
    echo "UNSWEPT $pr $branch"
    continue
  fi
  mg=$(gh pr view $pr --json mergeable -q .mergeable); [ "$mg" = CONFLICTING ] && { echo "CONFLICT $pr"; continue; }
  j=$(gh pr checks $pr --json bucket,name 2>/dev/null)
  pend=$(jq '[.[]|select(.bucket=="pending")]|length' <<<"$j"); fail=$(jq -r '[.[]|select(.bucket=="fail")|.name]|join(",")' <<<"$j")
  if jq -e '[.[]|select(.bucket=="pending" and (.name|test("^codecov/(patch|project)$")))]|length>0' <<<"$j" >/dev/null; then
    committed=$(gh pr view "$pr" --json commits -q '.commits[-1].committedDate')
    committed_epoch=$(date -d "$committed" +%s 2>/dev/null || echo 0)
    age_seconds=$(($(date +%s) - committed_epoch))
    if (( age_seconds >= 7200 )); then
      echo "STALE-CODECOV $pr $((age_seconds / 3600))h"
    fi
  fi
  if [ -n "$fail" ]; then echo "FAIL $pr: $fail"; continue; fi
  if ! grep -qx $pr "$BATCH_STATE/reviewed.txt"; then
    n=$(gh pr view $pr --json comments -q '[.comments[]|select(.author.login=="github-actions")]|length')
    [ "$n" -gt 0 ] && echo "REVIEW-READY $pr"; continue
  fi
  [ "$pend" = 0 ] && "$B/land.sh" "$pr"
done < <(jq -r '.[]|[.number,.headRefName]|@tsv' <<<"$prs")
