#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: land.sh <pr> — gate (with expected SHA = remote branch tip) then squash-merge. Caller must have read reviews.
cd "$REPO"
br=$(gh pr view $1 --json headRefName -q .headRefName)
git fetch -q origin "$br"; sha=$(git rev-parse origin/$br)
out=$(bash .claude/skills/land-pr/scripts/merge-gate.sh $1 $sha 2>&1)
echo "$out" | grep -q 'VERDICT: gate passed' || { echo "$out" | grep -vE '^\s+ok'; exit 1; }
gh pr view $1 --json comments -q '[.comments[]|select(.author.login=="codecov")]|length' | grep -q '^0$' && gh pr checks $1 --json name -q '.[].name' | grep -q codecov/patch || true
gh pr merge $1 --squash --delete-branch >/dev/null 2>&1
st=$(gh pr view $1 --json state -q .state); echo "PR $1 $st"
iss=$(gh pr view $1 --json closingIssuesReferences -q '.closingIssuesReferences[].number')
for i in $iss; do echo "  issue $i $(gh issue view $i --json state -q .state)"; done
git branch -D "$br" >/dev/null 2>&1; true
