#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# usage: fieldcheck.sh <worktree> — warn when added page state fields have no controller JSON field
cd "$1" || exit 0
controller_dir=internal/adapter/controller
mapfile -t changed_pages < <({
  git diff --name-only origin/develop...HEAD -- 'frontend/src/pages/*Page.tsx' 2>/dev/null
  git diff --name-only HEAD -- 'frontend/src/pages/*Page.tsx'
  git ls-files --others --exclude-standard -- 'frontend/src/pages/*Page.tsx'
} | sort -u)
for f in "${changed_pages[@]}"; do
  page=$(basename "$f" Page.tsx)
  ctrl=$(find "$controller_dir" -maxdepth 1 -type f -iname "${page}WebController.go" ! -name '*_test.go' -print)
  if [[ -z "$ctrl" ]]; then ctrl=$(find "$controller_dir" -maxdepth 1 -type f -name '*.go' ! -name '*_test.go' -print); fi
  output_bases=$(find "$controller_dir" -maxdepth 1 -type f -iname '*output_base*.go' ! -name '*_test.go' -print)
  files="$ctrl $output_bases"
  while IFS= read -r fld; do
    [[ -n "$fld" ]] || continue
    if ! grep -q "json:\"$fld[,\"]" $files 2>/dev/null; then
      echo "FIELDCHECK WARN: state.$fld read by the page, not in $files"
    fi
  done < <({
    git diff -U0 origin/develop...HEAD -- "$f" 2>/dev/null
    git diff -U0 HEAD -- "$f"
    if [[ -f "$f" ]] && ! git ls-files --error-unmatch "$f" >/dev/null 2>&1; then sed 's/^/+/' "$f"; fi
  } | grep '^+' | grep -v '^+++' | grep -oE '\bstate\??\.[a-zA-Z0-9_]+' | sed -E 's/state\??\.//')
done
