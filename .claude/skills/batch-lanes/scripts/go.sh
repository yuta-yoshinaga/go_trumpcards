#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: go.sh <slot> <issue#> <prompt-file> — delegate using the exact prompt file
prompt="${3:-${BATCH_PROMPT_FILE:-}}"
if [[ -z "$prompt" || ! -s "$prompt" ]]; then echo "Prompt file is required and must be non-empty (third argument or BATCH_PROMPT_FILE)" >&2; exit 2; fi
wt="$WT_ROOT/wt-ib$1"
[[ -d "$wt" ]] || { echo "Missing worktree: $wt" >&2; exit 2; }
run="$(date +%Y%m%d%H%M%S)-$$"
log="$BATCH_STATE/go-$1-$run.log"
result="$BATCH_STATE/r$2.txt"
if (cd "$wt" && dele -k edit -w -d "$wt" "$(cat "$prompt")") >"$log" 2>&1; then rc=0; else rc=$?; fi
cp "$log" "$result"
echo "DONE slot=$1 issue=$2 exit=$rc log=$log"
exit "$rc"
