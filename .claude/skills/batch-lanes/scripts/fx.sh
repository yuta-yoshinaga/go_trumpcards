#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: fx.sh <numeric-slot> <branch> — check out an existing PR branch in a fixup worktree; slot must be numeric (use a lane-distinct number such as 21..29)
[[ "${1:-}" =~ ^[0-9]+$ ]] || { echo "fx slot must be numeric (rcg/e2e use port 18000+slot); use e.g. 21 for a fixup lane" >&2; exit 2; }
wt="$WT_ROOT/wt-ib$1"; cd "$REPO"; git fetch -q origin "$2"
[ -d "$wt" ] || git worktree add -q --detach "$wt" origin/develop
cd "$wt"; test -z "$(git status --porcelain)" || { echo DIRTY; exit 1; }
while IFS= read -r other; do
  [ -n "$other" ] || continue
  [ "$other" = "$wt" ] && continue
  if [ "$(git -C "$other" symbolic-ref -q HEAD || true)" = "refs/heads/$2" ]; then
    if [ -n "$(git -C "$other" status --porcelain)" ]; then
      echo "DIRTY worktree: $other" >&2
      exit 1
    fi
    git -C "$other" checkout -q --detach || exit $?
  fi
done < <(git worktree list --porcelain | sed -n 's/^worktree //p')
git checkout -q -B "$2" "origin/$2" && git branch --set-upstream-to=origin/$2 -q
[ -d frontend/node_modules ] || (cd frontend && bun install --frozen-lockfile >/dev/null 2>&1)
echo "$wt $(git log --oneline -1)"
