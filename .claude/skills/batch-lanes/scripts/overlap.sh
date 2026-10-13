#!/usr/bin/env bash
set -euo pipefail
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"

slot="${1:?usage: overlap.sh <slot>}"
mine="$WT_ROOT/wt-ib$slot"
[[ -d "$mine" ]] || exit 0

gh_available=1
if command -v gh >/dev/null 2>&1; then
  if open_branches="$(gh pr list --state open --limit 200 --json headRefName --jq '.[].headRefName' 2>/dev/null)"; then
    :
  else
    gh_available=0
    open_branches=""
  fi
else
  gh_available=0
  open_branches=""
fi
if (( ! gh_available )); then
  printf 'NOTE LANE_OVERLAP gh unavailable; committed lanes not checked\n'
fi

collect() {
  local wt="$1"
  {
    git -C "$wt" diff --name-only -z origin/develop...HEAD 2>/dev/null || true
    git -C "$wt" diff --name-only -z HEAD 2>/dev/null || true
    git -C "$wt" ls-files --others --exclude-standard -z 2>/dev/null || true
  } | sort -zu
}

mapfile -d '' mine_files < <(collect "$mine")
shopt -s nullglob
for other in "$WT_ROOT"/wt-ib*; do
  [[ -d "$other" && "$other" != "$mine" ]] || continue
  other_slot="${other##*/wt-ib}"
  git -C "$other" symbolic-ref -q HEAD >/dev/null || continue
  dirty="$(git -C "$other" status --porcelain --untracked-files=all 2>/dev/null || true)"
  active=0
  [[ -n "$dirty" ]] && active=1
  if (( gh_available )) && grep -Fxq "$(git -C "$other" branch --show-current 2>/dev/null || true)" <<<"$open_branches"; then
    active=1
  fi
  (( active )) || continue
  mapfile -d '' other_files < <(collect "$other")
  declare -A theirs=()
  for file in "${other_files[@]}"; do theirs["$file"]=1; done
  for file in "${mine_files[@]}"; do
    [[ -n "${theirs[$file]+x}" ]] || continue
    case " ${ALLOW_OVERLAP:-} " in *" $file "*) continue ;; esac
    printf 'WARN LANE_OVERLAP %s slots=%s,%s\n' "$file" "$slot" "$other_slot"
  done
  unset theirs
done
