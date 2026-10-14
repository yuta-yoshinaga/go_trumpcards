#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

slots=()
if (( $# )); then
  for slot in "$@"; do
    [[ "$slot" =~ ^[0-9]+$ ]] || { echo "slot must be numeric: $slot" >&2; exit 2; }
    slots+=("$slot")
  done
else
  for ((slot=0; slot<=99; slot++)); do slots+=("$slot"); done
fi

found=0
for slot in "${slots[@]}"; do
  port=$((18000 + 10#$slot))
  if ! ss_output=$(ss -ltnpH "sport = :$port" 2>/dev/null); then
    echo "ss failed while checking port $port" >&2
    exit 1
  fi
  while IFS= read -r pid; do
    [[ "$pid" =~ ^[0-9]+$ ]] || continue
    [[ -r "/proc/$pid/comm" ]] || continue
    comm=$(<"/proc/$pid/comm")
    kill -TERM "$pid" 2>/dev/null || continue
    found=1
    for ((waited=0; waited<50; waited++)); do
      kill -0 "$pid" 2>/dev/null || break
      sleep 0.1
    done
    kill -0 "$pid" 2>/dev/null && kill -KILL "$pid" 2>/dev/null || true
    printf 'stopped port=%s pid=%s cmd=%s\n' "$port" "$pid" "$comm"
  done < <(printf '%s\n' "$ss_output" | grep -oE 'pid=[0-9]+' | cut -d= -f2 | sort -u)
done
(( found )) || echo none
