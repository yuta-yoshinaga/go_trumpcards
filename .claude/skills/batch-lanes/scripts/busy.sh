#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# busy.sh <slot> — exit 0 (and print PIDs) if a delegation process is writing in wt-ib<slot>.
# Reads /proc/<pid>/cmdline of codex/agy workers only, so it never matches the caller's own shell.
wt="$WT_ROOT/wt-ib$1"
for p in $(pgrep -x codex 2>/dev/null; pgrep -f "codex-task|agy-task" 2>/dev/null); do
  c=$( { tr "\0" " " < /proc/$p/cmdline; } 2>/dev/null) || continue
  case "$c" in *" -C $wt "*|*" -d $wt "*) echo "$p"; found=1;; esac
done
[ -n "$found" ]
