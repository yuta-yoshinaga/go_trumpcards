#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: fieldcheck.sh <worktree> — warn when added page lines read state.X that the game's controller does not send
cd "$1" || exit 0
game=$(git branch --show-current | sed -E 's#^feat/[0-9]+-##')
ctrl=$(ls internal/adapter/controller/ | grep -i "^${game}.*WebController.go$" | grep -v _test | sed 's#^#internal/adapter/controller/#')
[ -z "$ctrl" ] && ctrl='internal/adapter/controller/*.go'
# Embedded output bases (e.g. SolitaireWebOutputBase, WebOutputBase) carry shared fields.
[ -n "$ctrl" ] && ctrl="$ctrl $(ls internal/adapter/controller/*output_base*.go 2>/dev/null | grep -v _test | tr '\n' ' ')"
for f in $(git diff --name-only -- 'frontend/src/pages/*Page.tsx'); do
  git diff -U0 -- "$f" | grep '^+' | grep -v '^+++' | grep -oE '\bstate\??\.[a-zA-Z0-9_]+' | sed -E 's/state\??\.//'
done | sort -u | while read -r fld; do
  grep -q "json:\"$fld[,\"]" $ctrl 2>/dev/null || echo "FIELDCHECK WARN: state.$fld read by the page, not in $ctrl"
done
