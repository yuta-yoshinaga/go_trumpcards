#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
slot="${1:?usage: e2e.sh <slot> [playwright args...]}"; shift
[[ "$slot" =~ ^[0-9]+$ ]] || { echo "invalid slot: $slot" >&2; exit 2; }
declare -a run_args=() run_starts=(0)
for arg in "$@"; do
  if [[ "$arg" == --- ]]; then run_starts+=("${#run_args[@]}"); else run_args+=("$arg"); fi
done
wt="$WT_ROOT/wt-ib$slot"
[[ -d "$wt/frontend" ]] || { echo "missing lane worktree: $wt" >&2; exit 2; }
port=$((18000 + slot))
# Two shared locks cap concurrent server builds and browsers at the measured memory budget.
while :; do
  for lock in 0 1; do
    eval "exec $((7 + lock))>\"$BATCH_STATE/e2e-$lock.lock\""
    if flock -n "$((7 + lock))"; then chosen=$lock; break 2; fi
  done
  sleep 1
done
cleanup() {
  if [[ -n "${server_pid:-}" ]]; then kill "$server_pid" 2>/dev/null || true; wait "$server_pid" 2>/dev/null || true; fi
}
trap cleanup EXIT INT TERM
if [[ "${E2E_SKIP_BUILD:-0}" != 1 ]] && ! python3 - "$port" <<'PY'
import socket, sys
s = socket.socket()
try: s.bind(("127.0.0.1", int(sys.argv[1])))
except OSError: sys.exit(1)
finally: s.close()
PY
then echo "E2E port $port is already in use" >&2; exit 1; fi
if [[ "${E2E_SKIP_BUILD:-0}" != 1 ]]; then
  (cd "$wt/frontend" && bun run build)
  server_bin="$BATCH_STATE/e2e-server-$slot"
  (cd "$wt" && go build -o "$server_bin" ./cmd/server)
  (cd "$wt" && PORT="$port" "$server_bin") >"$BATCH_STATE/e2e-server-$slot.log" 2>&1 &
  server_pid=$!
  ready=0
  for _ in $(seq 60); do
    if curl --max-time 2 -fs "http://localhost:$port/" >/dev/null; then ready=1; break; fi
    kill -0 "$server_pid" 2>/dev/null || { cat "$BATCH_STATE/e2e-server-$slot.log" >&2; exit 1; }
    sleep 1
  done
  [[ $ready -eq 1 ]] || { echo "server did not become ready on port $port" >&2; exit 1; }
fi
status=0
for i in "${!run_starts[@]}"; do
  start="${run_starts[$i]}"; end="${run_starts[$((i + 1))]:-${#run_args[@]}}"
  args=("${run_args[@]:start:end-start}")
  echo "run $((i + 1)): ${args[*]}"
  (cd "$wt/frontend" && E2E_PORT="$port" E2E_EXTERNAL_SERVER=1 bunx playwright test --reporter=line "${args[@]}") || status=1
done
exit "$status"
