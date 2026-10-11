#!/usr/bin/env bash
set -euo pipefail
for command in git awk grep python3; do
  command -v "$command" >/dev/null 2>&1 || { echo "FAIL: required command not found: $command" >&2; exit 127; }
done
if (($# < 2)); then echo "usage: after.sh OLD NEW [file ...]" >&2; exit 2; fi
old=$1 new=$2; shift 2
files=("$@")
if ((${#files[@]} == 0)); then
  mapfile -t files < <(git diff HEAD --name-only -- frontend/src/i18n/locales internal/i18n/locales)
fi
((${#files[@]})) || { echo 'FAIL: no locale files in scope' >&2; exit 1; }
failed=0
# Changed locale files must preserve line count (numstat additions == deletions).
while IFS=$'\t' read -r added deleted path; do
  [[ -n "$path" ]] || continue
  if [[ "$added" != "$deleted" ]]; then echo "FAIL: line count changed: $path (+$added -$deleted)" >&2; failed=1; fi
done < <(git diff HEAD --numstat -- "${files[@]}")
# Repetition check on Japanese locale files in the selected scope.
for file in "${files[@]}"; do
  [[ "$file" == */ja/*.json ]] || continue
  if ! python3 - "$file" <<'PY'
import re, sys
pattern = re.compile(r'(\S+)（\1）')
found = False
with open(sys.argv[1], encoding='utf-8') as source:
    for number, line in enumerate(source, 1):
        if pattern.search(line):
            print(f'{sys.argv[1]}:{number}:{line.rstrip()}')
            found = True
sys.exit(1 if found else 0)
PY
  then echo "FAIL: repeated Japanese term in $file" >&2; failed=1; fi
done
count() { (git grep -F -c -- "$1" HEAD -- "${files[@]}" 2>/dev/null || true) | awk -F: '{n += $NF} END {print n+0}'; }
old_count=$(count "$old"); new_count=$(count "$new"); wrapped_count=$(count "$new（")
# Count current working-tree occurrences; require old term gone and new term present.
current_count() { (grep -F -c -- "$1" "${files[@]}" 2>/dev/null || true) | awk -F: '{n += $NF} END {print n+0}'; }
old_now=$(current_count "$old"); new_now=$(current_count "$new"); wrapped_now=$(current_count "$new（")
if ((old_now != 0)); then echo "FAIL: old term remains ($old_now; HEAD $old_count)" >&2; failed=1; fi
if ((new_now < new_count)); then echo "FAIL: new term count decreased ($new_now; HEAD $new_count)" >&2; failed=1; fi
if ((wrapped_now > wrapped_count)); then echo "FAIL: new（ count increased ($wrapped_now; HEAD $wrapped_count)" >&2; failed=1; fi
((failed == 0)) || exit 1
echo "PASS: saved line counts; old=0, new=$new_now (HEAD $new_count), new（=$wrapped_now (HEAD $wrapped_count); no repeated ja terms"
