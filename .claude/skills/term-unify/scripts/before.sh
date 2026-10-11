#!/usr/bin/env bash
set -euo pipefail
if (($# < 2)); then echo "usage: before.sh OLD NEW [file ...]" >&2; exit 2; fi
old=$1 new=$2; shift 2
files=("$@")
if ((${#files[@]} == 0)); then files=(frontend/src/i18n/locales internal/i18n/locales); fi
count() { (git grep -F -c -- "$1" HEAD -- "${files[@]}" 2>/dev/null || true) | awk -F: '{n += $NF} END {print n+0}'; }
printf 'HEAD old (%s): %s\n' "$old" "$(count "$old")"
printf 'HEAD new (%s): %s\n' "$new" "$(count "$new")"
printf 'HEAD new（ (%s（): %s\n' "$new" "$(count "$new（")"
