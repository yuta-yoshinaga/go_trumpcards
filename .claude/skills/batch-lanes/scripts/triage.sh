#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# triage.sh — for unreviewed open PRs with a finished review: auto-mark clean ones, print flagged lines
require_branch_re || exit $?
branch_re="$BATCH_BRANCH_RE"
for pr in $(gh pr list --state open --author @me --limit 200 --json number,headRefName | jq -r --arg re "$branch_re" '.[]|select(.headRefName|test($re))|.number' | sort -n); do
  grep -qx $pr "$BATCH_STATE/reviewed.txt" && continue
  b=$(gh pr view $pr --json comments -q '[.comments[]|select(.author.login=="github-actions")|.body]|last')
  grep -q finished <<<"$b" || continue
  f=$(sed -n '/^---/,$p' <<<"$b" | grep -iE "must fix|should fix|should consider|medium|please|robust|regress|bug|incorrect|wrong|leak|stale|mismatch|unreachable|silent|hidden|hides|never re|plural" | grep -viE "must fix[^a-z]*(none|nothing)|nothing (that )?must|no (must|blocking)|not a bug|no bug|is never|never leaves|stays silent|silent on (the )?(initial|first)|won.t leak|can.t leak|doesn.t leak|not leak|no leak|plural (forms|resolution|keys)|_one|plural.*(correct|fine|right)" | cut -c1-240 | head -3)
  if [ -z "$f" ]; then echo $pr >> "$BATCH_STATE/reviewed.txt"; echo "OK $pr"; else echo "=== $pr"; echo "$f"; fi
done
