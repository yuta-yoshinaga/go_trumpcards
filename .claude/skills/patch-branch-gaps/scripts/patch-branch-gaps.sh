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
  python3 - "$out/coverage-final.json" "$src" "$TMP/diff" <<'PY'
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
  [[ $rc -eq 2 ]] && gaps=1
  # Python emits rows but does not signal their presence; detect them from a second focused pass.
  if python3 - "$out/coverage-final.json" "$src" "$TMP/diff" <<'PY'
import json,re,sys
d=json.load(open(sys.argv[1])); p=sys.argv[2]; changed=set()
for x in open(sys.argv[3]):
 m=re.match(r'@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@',x)
 if m: changed.update(range(int(m[1]),int(m[1])+int(m[2] or 1)))
k=next((k for k in d if k.endswith('/'+p.removeprefix('frontend/')) or k.endswith(p)),None)
if k and any(b['loc']['start']['line'] in changed and any(n==0 for n in d[k]['b'][i]) for i,b in d[k].get('branchMap',{}).items()): sys.exit(0)
sys.exit(1)
PY
  then gaps=1; fi
done <<< "$sources"
if [[ $gaps -eq 0 ]]; then echo "patch-branch-gaps: PASS (no uncovered branches on changed lines)"; else echo "patch-branch-gaps: gaps found or measurement failed."; fi
exit "$gaps"
