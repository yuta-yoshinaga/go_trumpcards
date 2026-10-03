#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: gofx.sh <fxslot> <pr-or-tag> <prompt-file>
# Refuses to delegate when the tag is a PR number and the worktree is not on that PR's head branch.
cd "$WT_ROOT/wt-ib$1" || exit 1
prompt="${3:-${BATCH_PROMPT_FILE:-}}"
if [[ -z "$prompt" || ! -s "$prompt" ]]; then echo "Prompt file is required and must be non-empty (third argument or BATCH_PROMPT_FILE)" >&2; exit 2; fi
if [[ "$2" =~ ^[0-9]+$ ]]; then
  want=$(gh pr view "$2" --json headRefName -q .headRefName 2>/dev/null)
  have=$(git branch --show-current)
  if [ -n "$want" ] && [ "$want" != "$have" ]; then echo "DONE fx=$1 pr=$2 exit=BRANCH_MISMATCH have=$have want=$want"; exit 1; fi
fi
run="$(date +%Y%m%d%H%M%S)-$$"; log="$BATCH_STATE/gofx-$1-$run.log"
if dele -k edit -w -d "$WT_ROOT/wt-ib$1" "$(cat "$prompt")" >"$log" 2>&1; then rc=0; else rc=$?; fi
cp "$log" "$BATCH_STATE/rf$2.txt"
echo "DONE fx=$1 pr=$2 exit=$rc log=$log"
exit "$rc"
