#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

slot="${1:?usage: adrcheck.sh <slot>}"
lane="$WT_ROOT/wt-ib$slot"
[[ -d "$lane" ]] || exit 0
[[ "${ALLOW_ADR_EDIT:-}" == 1 ]] && exit 0

mapfile -d '' -t changed < <(collect "$lane")
for file in "${changed[@]}"; do
  [[ "$file" == docs/adr/*.md ]] || continue
  printf 'WARN ADR_EDIT %s\n' "$file"
done
