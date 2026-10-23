#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: setup_wt.sh <slot> <issue#> <game>
set -e
slot="$1" n="$2" g="$3"
wt="$WT_ROOT/wt-ib$slot"
if [[ "${ALLOW_NON_GAME_BRANCH:-0}" != 1 ]] && ! batch_branch_has_game "$REPO" "feat/$n-$g"; then
  echo "invalid game in branch name: $g (set ALLOW_NON_GAME_BRANCH=1 to bypass)" >&2
  exit 2
fi
cd "$REPO"
for i in 1 2 3 4 5; do git fetch -q origin develop && break; sleep 3; done
if [ -d "$wt" ]; then
  [ -f "$wt/-" ] && [ ! -s "$wt/-" ] && rm -f "$wt/-"
  cd "$wt"; test -z "$(git status --porcelain)" || { echo "DIRTY $wt"; exit 1; }
  git checkout -q --detach origin/develop
else
  git worktree add -q --detach "$wt" origin/develop; cd "$wt"
fi
git checkout -q -b "feat/$n-$g"
printf '%s\n' "$(git branch --show-current)" > "$BATCH_STATE/slot$slot.branch"
[ -d frontend/node_modules ] || (cd frontend && bun install --frozen-lockfile >/dev/null 2>&1)
echo "$wt $(git branch --show-current)"
