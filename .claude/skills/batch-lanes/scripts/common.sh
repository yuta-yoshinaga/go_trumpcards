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
