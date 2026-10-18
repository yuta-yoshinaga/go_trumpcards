#!/usr/bin/env bash
# Shared path configuration for batch-lanes scripts.
B="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="${BATCH_REPO:-$(git -C "$B/../../../../" rev-parse --show-toplevel)}"
REPO="$(cd "$REPO" && pwd)"
COMMON_DIR="$(git -C "$REPO" rev-parse --git-common-dir)"
[[ "$COMMON_DIR" = /* ]] || COMMON_DIR="$REPO/$COMMON_DIR"
BATCH_STATE="${BATCH_STATE:-$COMMON_DIR/batch-lanes}"
BATCH_WT_ROOT="${BATCH_WT_ROOT:-$(dirname "$REPO")}"
WT_ROOT="$BATCH_WT_ROOT"
mkdir -p "$BATCH_STATE"
touch "$BATCH_STATE/reviewed.txt" "$BATCH_STATE/acked.txt"

require_branch_re() {
  if [[ -z "${BATCH_BRANCH_RE:-}" ]]; then
    echo "Set BATCH_BRANCH_RE to select batch branches (for example: ^feat/[0-9]+-)" >&2
    return 2
  fi
}

delegated_to() {
  local slot="$1" log backend branch_file current_branch
  if [[ -n "${BATCH_DELEGATED_TO:-}" ]]; then printf '%s' "$BATCH_DELEGATED_TO"; return; fi
  branch_file="$BATCH_STATE/slot$slot.branch"
  current_branch=$(git -C "$WT_ROOT/wt-ib$slot" branch --show-current 2>/dev/null || true)
  if [[ ! -f "$branch_file" || -z "$current_branch" || "$(cat "$branch_file")" != "$current_branch" ]]; then
    printf 'unknown'
    return
  fi
  local -a backends=()
  shopt -s nullglob
  for log in "$BATCH_STATE"/go-"$slot"-*.log; do
    [[ "$log" -nt "$branch_file" ]] || continue
    while IFS= read -r backend; do [[ -n "$backend" ]] && backends+=("$backend"); done \
      < <(sed -nE 's/.*dele: ([^[:space:]]+) ok.*/\1/p' "$log")
  done
  shopt -u nullglob
  if ((${#backends[@]} == 0)); then printf 'none (orchestrator)'; else
    local -a unique=()
    mapfile -t unique < <(printf '%s\n' "${backends[@]}" | LC_ALL=C sort -u)
    local joined='' item
    for item in "${unique[@]}"; do joined+="${joined:+, }$item"; done
    printf '%s' "$joined"
  fi
}
