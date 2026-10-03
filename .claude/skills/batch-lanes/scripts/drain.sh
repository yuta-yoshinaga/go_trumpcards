#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# drain.sh — sweep every 5 min; exit on FAIL/CONFLICT/new review or when no batch PR is open
require_branch_re || exit $?
branch_re="$BATCH_BRANCH_RE"
cd "$REPO"
for i in $(seq 1 30); do
  out=$("$B/sweep.sh" 2>&1); echo "$(date +%H:%M) $out" | grep -v '^\S* $'
  echo "$out" | grep -qE 'FAIL|CONFLICT' && exit 0
  n=$(gh pr list --state open --author @me --json headRefName | jq --arg re "$branch_re" '[.[]|select(.headRefName|test($re))]|length')
  [ "$n" = 0 ] && { echo "ALL MERGED"; exit 0; }
  timeout 300 tail -f /dev/null
done
