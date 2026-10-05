#!/usr/bin/env bash
set -uo pipefail

# 限界: タグ無しの共有ファイルだけを変えた場合は、それに依存するタグ付きファイルの worker を型検査しない。

# Print worker build tags from a Go build constraint on the first line.
worker_tags_from_line() {
  local line="${1:-}"
  [[ "$line" == '//go:build '* ]] || return 0
  printf '%s\n' "${line#//go:build }" | grep -oE '[[:alnum:]_]+' | grep -Ev '^(js|wasm)$' | sort -u || true
}

if [[ "${BATCH_TESTING:-}" == 1 ]]; then return 0 2>/dev/null || exit 0; fi
worktree="${1:?usage: workerbuild.sh <worktree>}"
cd "$worktree" || exit 2
mapfile -d '' changed < <(git diff --name-only -z HEAD; git ls-files --others --exclude-standard -z)
declare -A tags=()
for file in "${changed[@]}"; do
  [[ "$file" == *.go && "$file" != *_test.go && -f "$file" ]] || continue
  IFS= read -r first < "$file" || true
  while IFS= read -r tag; do [[ -n "$tag" ]] && tags["$tag"]=1; done < <(worker_tags_from_line "$first")
done
for tag in "${!tags[@]}"; do
  [[ -d "cmd/workers/$tag" ]] || continue
  output="$(GOOS=js GOARCH=wasm go build -tags "$tag" -o /dev/null "./cmd/workers/$tag" 2>&1)" || {
    first_error="$(printf '%s\n' "$output" | sed -n '1p')"
    printf 'WARN WORKER_BUILD %s: %s\n' "$tag" "$first_error"
  }
done
