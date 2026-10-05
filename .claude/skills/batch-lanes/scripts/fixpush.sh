#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# refuse while a delegation is still writing in this worktree (2026-10-01: a re-run shipped #9574's half-written files as #9564)
$B/busy.sh "$1" >/dev/null && { echo "REFUSED: a delegation is still running in wt-ib$1"; exit 1; }
# refuse unless rcg.sh passed on exactly this worktree state
[ "$(cat "$BATCH_STATE/gate$1.ok" 2>/dev/null)" = "$("$B/treehash.sh" "$1")" ] || { echo "REFUSED: no passing rcg.sh for slot $1 at this worktree state"; exit 1; }
# usage: fixpush.sh <fxslot> <pr> "<commit subject>" "<reply to review>"
set -e
cd "$WT_ROOT/wt-ib$1"
if [[ "${CLAIM_SKIP:-}" != 1 ]] && ! "$B/claimcheck.sh" "$PWD" "$4"; then
  echo "REFUSED: unverified claims" >&2
  exit 1
fi
mapfile -d '' -t files < <({ git diff --name-only -z HEAD; git ls-files -o --exclude-standard -z; } | sort -zu)
if ((${#files[@]})); then git add -- "${files[@]}"; fi
git commit -q -F - <<M
$3

Delegated-To: ${BATCH_DELEGATED_TO:-codex (dele -k edit)}
${CLAUDE_CO_AUTHOR:+Co-Authored-By: $CLAUDE_CO_AUTHOR}
${CLAUDE_SESSION_URL:+Claude-Session: $CLAUDE_SESSION_URL}
M
set +e
git push -q 2>&1 | grep -v '^remote:'
push_status=${PIPESTATUS[0]}
set -e
((push_status == 0)) || exit "$push_status"
sha=$(git rev-parse --short HEAD); rsha=$(git ls-remote origin "refs/heads/$(git branch --show-current)" | cut -c1-9)
[ "${sha:0:9}" = "$rsha" ] || { echo "PUSH MISMATCH $sha $rsha"; exit 1; }
tn=$(git show HEAD -U0 -- "*.test.ts" "*.test.tsx" "*_test.go" | grep -E "^\+\s*(it|test)\(|^\+func Test" | sed -E "s/^\+\s*//; s/, async.*//; s/\{$//" | head -8)
[ -n "$tn" ] && body="$4

Tests added/changed in this commit (from the diff):
$(sed "s/^/- /" <<<"$tn")" || body="$4"
gh pr comment "$2" --body "$body ($sha)" >/dev/null && echo "pushed $sha + replied"
