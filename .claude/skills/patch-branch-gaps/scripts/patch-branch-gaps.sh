#!/usr/bin/env bash
# Find zero-hit frontend branches on changed lines, using each source file's sibling test.
set -uo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null)" || { echo "patch-branch-gaps: run inside a git worktree" >&2; exit 2; }
cd "$ROOT"
BASE="${1:-$(git merge-base origin/develop HEAD 2>/dev/null)}"
if [[ -z "$BASE" ]]; then
  echo "patch-branch-gaps: cannot determine base; pass a base ref" >&2
  exit 2
fi
TMP="$(mktemp -d)" || exit 2
trap 'rm -rf "$TMP"' EXIT
changed="$(git diff --name-only "$BASE" -- frontend/src/{pages,components,utils,hooks,api} 2>/dev/null | sort -u)"
sources="$(printf '%s\n' "$changed" | grep -E '^frontend/src/(pages|components|utils|hooks|api)/.*\.tsx?$' | grep -v '\.test\.tsx?$' || true)"
if [[ -z "$sources" ]]; then
  echo "patch-branch-gaps: no changed frontend source files under frontend/src/{pages,components,utils,hooks,api}."
  exit 0
fi

gaps=0
new_gaps=0
preexisting_gaps=0
while IFS= read -r src; do
  [[ -z "$src" ]] && continue
  rel="${src#frontend/}"
  ext="${src##*.}"
  test_file="${rel%.*}.test.${ext}"
  if [[ ! -f "frontend/$test_file" ]]; then
    printf '%s: NO TEST FILE\n' "$rel"
    continue
  fi
  out="$TMP/coverage"
  rm -rf "$out"
  if ! (cd frontend && bunx vitest run "$test_file" --coverage --coverage.include="$rel" --coverage.reporter=json --coverage.reportsDirectory="$out") >"$TMP/vitest.log" 2>&1; then
    printf '%s: NO RESULT (vitest failed; see %s)\n' "$rel" "$TMP/vitest.log" >&2
    cat "$TMP/vitest.log" >&2
    gaps=1
    continue
  fi
  if [[ ! -f "$out/coverage-final.json" ]]; then
    printf '%s: NO RESULT (coverage-final.json missing)\n' "$rel" >&2
    gaps=1
    continue
  fi
  git diff -U0 "$BASE" -- "$src" > "$TMP/diff"
  python3 - "$out/coverage-final.json" "$src" "$TMP/diff" >"$TMP/row" <<'PY'
import json, re, sys
coverage_path, source_path, diff_path = sys.argv[1:]
changed = set()
for line in open(diff_path):
    m = re.match(r'@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@', line)
    if m:
        start, count = int(m.group(1)), int(m.group(2) or 1)
        changed.update(range(start, start + count))
doc = json.load(open(coverage_path))
key = next((k for k in doc if k.endswith('/' + source_path.removeprefix('frontend/')) or k.endswith(source_path)), None)
if key is None:
    print(f'{source_path}: NO RESULT (source absent from coverage report)', file=sys.stderr)
    sys.exit(2)
cov = doc[key]
src = open(source_path).read().splitlines()
for bid, branch in cov.get('branchMap', {}).items():
    ln = branch['loc']['start']['line']
    counts = cov['b'][bid]
    if ln in changed and any(n == 0 for n in counts):
        print(f"{source_path.removeprefix('frontend/')}:{ln}: {branch['type']} {counts} | {src[ln-1].strip()[:110]}")
PY
  rc=$?
  [[ $rc -ne 0 ]] && gaps=1
  if [[ $rc -eq 0 ]]; then
    while IFS= read -r line; do
      [[ -z "$line" ]] && continue
      gap_file="${line%%:*}"
      gap_text="${line#* | }"
      gap_text="$(sed 's/^[[:space:]]*//; s/[[:space:]]*$//' <<< "$gap_text")"
      develop_file="$(git show "$BASE:frontend/$gap_file" 2>/dev/null | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')" || develop_file=""
      preexisting=0
      if [[ -n "$develop_file" && ${#gap_text} -ge 8 ]]; then
        while IFS= read -r develop_line; do
          [[ "${develop_line:0:${#gap_text}}" == "$gap_text" ]] && { preexisting=1; break; }
        done <<< "$develop_file"
      fi
      if [[ $preexisting -eq 1 ]]; then
        printf 'PREEXISTING %s\n' "$line"
        preexisting_gaps=$((preexisting_gaps + 1))
      else
        printf '%s\n' "$line"
        new_gaps=$((new_gaps + 1))
        gaps=1
      fi
    done < "$TMP/row"
  fi
done <<< "$sources"
echo "patch-branch-gaps: summary: $new_gaps new, $preexisting_gaps pre-existing"
if [[ $gaps -eq 0 ]]; then echo "patch-branch-gaps: PASS (no new uncovered branches on changed lines)"; else echo "patch-branch-gaps: new gaps found or measurement failed."; fi
exit "$gaps"
