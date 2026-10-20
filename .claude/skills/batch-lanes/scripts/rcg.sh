#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# rcg.sh <slot> — run the receipt and exit 0 only when every gate passed.
# On success it stamps gate<slot>.ok with the worktree hash; ship.sh/shipp.sh/fixpush.sh refuse without a matching stamp.
slot="$1"
pidfile="$BATCH_STATE/rcg$slot.pid"
if [[ -s "$pidfile" ]]; then
  old_pid=$(cat "$pidfile")
  stop_tree() {
    local parent="$1" child
    while read -r child; do [[ -n "$child" ]] && stop_tree "$child"; done < <(pgrep -P "$parent" 2>/dev/null || true)
    kill "$parent" 2>/dev/null || true
  }
  if rcg_live "$slot" "$old_pid"; then
    stop_tree "$old_pid"
  fi
fi
printf '%s\n' "$$" > "$pidfile"
started="$(date +%s%N)"
printf '%s\n' "$started" > "$BATCH_STATE/rcg$slot.started"
exec 9>"$BATCH_STATE/receipt.lock"
flock -x 9
rm -f "$BATCH_STATE/gate$slot.ok"
# Report develop changes that overlap this lane; NOTE lines are informational to the gate.
while IFS= read -r file; do [[ -n "$file" ]] && echo "NOTE STALE_BASE $file (run sync.sh $slot)"; done < <(stale_base_files "$WT_ROOT/wt-ib$slot")
"$B/rc.sh" "$slot" >/dev/null
rc=$?
f="$BATCH_STATE/rc$slot.txt"
printf '%s %s\n' "$started" "$(date +%s%N)" > "$BATCH_STATE/rcg$slot.result"
[[ $rc -eq 0 ]] || { echo "GATE: receipt failed"; exit "$rc"; }
grep -E "==|FAIL|WARN|NOTE|CHECK|TYPE|Tests |LINT|GOIMPORTS|MUTANT|MISSING" $f
grep -q '^CHECK=ok' $f || { echo "GATE: check failed"; exit 1; }
grep -q '^TYPE=ok' $f || { echo "GATE: typecheck failed"; exit 1; }
if grep -q '^LINT=FAIL' "$f"; then echo "GATE: lint failed"; exit 1; fi
if grep -q '^LINT' "$f" && ! grep -q '^LINT=ok' "$f"; then echo "GATE: lint unverified"; exit 1; fi
# NOTE lines quote source text, which may itself contain FAIL or WARN; they never fail the gate.
grep -v '^NOTE ' $f | grep -qE 'FAIL|MUTANT_LEFT|MISSING_TEST|GOIMPORTS:|LINT: [1-9]|WARN' && { echo "GATE: failure/warning present"; exit 1; }
if [[ "${E2E_GATE:-1}" == 0 ]]; then
  echo "NOTE E2E_GATE skipped"
else
  gate_hash="$("$B/treehash.sh" "$slot")"
  flock -u 9
  "$B/e2egate.sh" "$slot" || { echo "GATE: e2e failed"; exit 1; }
  flock -x 9
  [[ "$gate_hash" == "$("$B/treehash.sh" "$slot")" ]] || { echo "GATE: worktree changed during e2e"; exit 1; }
  printf '%s\n' "$gate_hash" > "$BATCH_STATE/gate$slot.ok"
fi
if [[ "${E2E_GATE:-1}" == 0 ]]; then "$B/treehash.sh" "$slot" > "$BATCH_STATE/gate$slot.ok"; fi
echo "GATE: ok"
