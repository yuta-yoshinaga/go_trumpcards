#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: land.sh <pr> — gate (with expected SHA = remote branch tip) then squash-merge. Caller must have read reviews.
cd "$REPO"
br=$(gh pr view "$1" --json headRefName -q .headRefName)
git fetch -q origin "$br"; sha=$(git rev-parse "origin/$br")
out=$(bash .claude/skills/land-pr/scripts/merge-gate.sh "$1" "$sha" 2>&1)
echo "$out" | grep -q 'VERDICT: gate passed' || { echo "$out" | grep -vE '^\s+ok'; exit 1; }
if gh pr merge "$1" --squash --delete-branch >/dev/null; then
  :
else
  rc=$?
  echo "Failed to merge PR $1" >&2
  exit "$rc"
fi
st=$(gh pr view "$1" --json state -q .state); echo "PR $1 $st"
iss=$(gh pr view "$1" --json closingIssuesReferences -q '.closingIssuesReferences[].number')
while IFS= read -r i; do [[ -n "$i" ]] && echo "  issue $i $(gh issue view "$i" --json state -q .state)"; done <<< "$iss"
git branch -D "$br" >/dev/null 2>&1; true
