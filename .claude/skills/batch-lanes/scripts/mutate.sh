#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
slot="${1:?usage: mutate.sh <slot> <ref>:<path> <e2e.sh args...>}"; shift
source_spec="${1:?usage: mutate.sh <slot> <ref>:<path> <e2e.sh args...>}"; shift
[[ "$source_spec" == *:* ]] || { echo "expected <ref>:<path>" >&2; exit 2; }
ref=${source_spec%%:*}; path=${source_spec#*:}
[[ -n "$ref" && -n "$path" ]] || { echo "expected <ref>:<path>" >&2; exit 2; }
wt="$WT_ROOT/wt-ib$slot"
target="$wt/$path"
[[ -f "$target" ]] || { echo "missing target: $target" >&2; exit 2; }
backup=$(mktemp "$BATCH_STATE/mutate.XXXXXX")
cp -p "$target" "$backup"
restored=0
restore() {
  if [[ "$restored" == 0 ]]; then cp -p "$backup" "$target"; restored=1; fi
}
cleanup() {
  restore
  rm -f "$backup"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
e2e_script="${BATCH_E2E_SCRIPT:-$B/e2e.sh}"
if ! bash "$e2e_script" "$slot" "$@"; then
  echo 'ERROR baseline failed' >&2
  exit 1
fi
git -C "$wt" show "$source_spec" > "$target"
if bash "$e2e_script" "$slot" "$@"; then
  result='NOT DETECTED'
else
  result='DETECTED'
fi
restore
if ! cmp -s "$target" "$backup"; then
  echo 'ERROR mutate.sh failed to restore target' >&2
  exit 1
fi
rm -f "$backup"
restored=1
trap - EXIT INT TERM
echo "$result"
