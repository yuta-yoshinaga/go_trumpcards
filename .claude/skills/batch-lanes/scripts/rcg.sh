#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# rcg.sh <slot> — run the receipt and exit 0 only when every gate passed.
# On success it stamps gate<slot>.ok with the worktree hash; ship.sh/shipp.sh/fixpush.sh refuse without a matching stamp.
slot="$1"
started="$(date +%s%N)"
printf '%s\n' "$started" > "$BATCH_STATE/rcg$slot.started"
exec 9>"$BATCH_STATE/receipt.lock"
flock -x 9
rm -f "$BATCH_STATE/gate$slot.ok"
"$B/rc.sh" "$slot" >/dev/null
rc=$?
f="$BATCH_STATE/rc$slot.txt"
printf '%s %s\n' "$started" "$(date +%s%N)" > "$BATCH_STATE/rcg$slot.result"
[[ $rc -eq 0 ]] || { echo "GATE: receipt failed"; exit "$rc"; }
grep -E "==|FAIL|WARN|CHECK|TYPE|Tests |LINT|GOIMPORTS|MUTANT|MISSING" $f
grep -q '^CHECK=ok' $f || { echo "GATE: check failed"; exit 1; }
grep -q '^TYPE=ok' $f || { echo "GATE: typecheck failed"; exit 1; }
grep -qE 'FAIL|MUTANT_LEFT|MISSING_TEST|GOIMPORTS:|LINT: [1-9]|WARN' $f && { echo "GATE: failure/warning present"; exit 1; }
"$B/treehash.sh" "$slot" > "$BATCH_STATE/gate$slot.ok"
echo "GATE: ok"
