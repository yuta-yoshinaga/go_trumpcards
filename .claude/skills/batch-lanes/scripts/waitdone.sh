#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# waitdone.sh [slot ...] — wait for rcg results newer than the corresponding start stamp
touch "$BATCH_STATE/acked.txt"
rcg_result_is_fresh() {
  local startfile="$1" result="$2" started result_started finished
  [[ -s "$startfile" && -s "$result" ]] || return 1
  read -r started < "$startfile"
  read -r result_started finished < "$result"
  [[ "$started" = "$result_started" ]] && (( finished > started ))
}
if [[ "${BATCH_TESTING:-}" = 1 ]]; then return 0; fi
while :; do
  found=0
  slots=("$@")
  if [[ ${#slots[@]} -eq 0 ]]; then
    for f in "$BATCH_STATE"/rcg*.result; do [[ -e "$f" ]] && slots+=("${f##*rcg}"); done
    for i in "${!slots[@]}"; do slots[$i]="${slots[$i]%.result}"; done
  fi
  for slot in "${slots[@]}"; do
    startfile="$BATCH_STATE/rcg$slot.started"; result="$BATCH_STATE/rcg$slot.result"
    rcg_result_is_fresh "$startfile" "$result" || continue
    read -r _ finished < "$result"
    key="rcg$slot $finished"
    grep -qxF "$key" "$BATCH_STATE/acked.txt" || { echo "$key" | tee -a "$BATCH_STATE/acked.txt"; found=1; }
  done
  [ $found = 1 ] && exit 0
  timeout 20 tail -f /dev/null
done
