#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
setup_repo() {
  rm -rf "$TMP/repo"
  mkdir -p "$TMP/repo/internal/i18n/locales/ja" "$TMP/repo/.git/objects/info"
  cd "$TMP/repo"
  git init -q
  git config user.email test@example.com
  git config user.name test
  # Reuse the caller's existing commit and objects as the baseline; never
  # create a commit in this disposable fixture repository.
  printf '%s\n' "$(git -C "$ROOT" rev-parse --path-format=absolute --git-path objects)" > .git/objects/info/alternates
  git update-ref HEAD "$(git -C "$ROOT" rev-parse HEAD)"
  cp "$ROOT/.claude/skills/term-unify/scripts/after.sh" after.sh
  cp "$ROOT/.claude/agents/locale-value-parity-auditor.md" auditor.md
  git -C "$ROOT" show HEAD:internal/i18n/locales/ja/solowhist.json > internal/i18n/locales/ja/solowhist.json
  git add .
}
write_62_lines() {
  printf '%s\n' "$1" > internal/i18n/locales/ja/solowhist.json
  for _ in {1..61}; do printf '\n' >> internal/i18n/locales/ja/solowhist.json; done
}
expect_failure() {
  local expected=$1 output status=0
  output=$(bash ./after.sh old new internal/i18n/locales/ja/solowhist.json 2>&1) || status=$?
  if ((status == 0)) || [[ "$output" != *"$expected"* ]]; then
    echo "expected after.sh failure containing '$expected'; status=$status output=$output" >&2
    exit 1
  fi
}
setup_repo
write_62_lines 'new（new）'
expect_failure 'FAIL: repeated Japanese term'
setup_repo
printf 'stable\n' > internal/i18n/locales/ja/solowhist.json
expect_failure 'FAIL: line count changed'
setup_repo
write_62_lines 'new term'
bash ./after.sh old new internal/i18n/locales/ja/solowhist.json

# A staged replacement still participates in line-count validation.
write_62_lines 'new term'
git add internal/i18n/locales/ja/solowhist.json
write_62_lines 'new term'
bash ./after.sh old new internal/i18n/locales/ja/solowhist.json

# Staging a line deletion must not hide it from after.sh.
printf 'new term\n' > internal/i18n/locales/ja/solowhist.json
git add internal/i18n/locales/ja/solowhist.json
if output=$(bash ./after.sh old new internal/i18n/locales/ja/solowhist.json 2>&1); then
  echo 'expected staged line deletion to fail after.sh' >&2
  exit 1
fi
[[ "$output" == *'FAIL: line count changed'* ]] || { echo "$output" >&2; exit 1; }

# The auditor should ignore roots where neither locale exists, while reporting
# a twin only when exactly one locale file exists.
mkdir -p frontend/src/i18n/locales/ja frontend/src/i18n/locales/en internal/i18n/locales/ja
printf '{"key":"日本語"}\n' > frontend/src/i18n/locales/ja/only.json
printf '{"key":"English"}\n' > frontend/src/i18n/locales/en/only.json
printf '{"key":"日本語"}\n' > internal/i18n/locales/ja/half.json
sed -n '/^python3 - "${FILES\[@\]}" <<'"'"'PY'"'"'$/,/^PY$/p' auditor.md | sed '1d;$d' > scanner.py
output=$(python3 scanner.py only.json half.json)
[[ "$output" != *'internal/i18n/locales/only.json::missing-twin:'* ]]
[[ "$output" == *'internal/i18n/locales/half.json::missing-twin:en'* ]]
printf 'term-unify script tests: PASS\n'
