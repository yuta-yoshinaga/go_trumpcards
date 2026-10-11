#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
setup_repo() {
  rm -rf "$TMP/repo"
  mkdir -p "$TMP/repo/internal/i18n/locales/ja"
  cd "$TMP/repo"
  git init -q
  git config user.email test@example.com
  git config user.name test
  cp "$ROOT/.claude/skills/term-unify/scripts/after.sh" after.sh
  cp "$ROOT/.claude/agents/locale-value-parity-auditor.md" auditor.md
  printf 'old term\nstable\n' > internal/i18n/locales/ja/game.json
  git add .
  printf '%s\n' "$(git write-tree)" > .git/HEAD
}
expect_failure() {
  local expected=$1 output status=0
  output=$(bash ./after.sh old new internal/i18n/locales/ja/game.json 2>&1) || status=$?
  if ((status == 0)) || [[ "$output" != *"$expected"* ]]; then
    echo "expected after.sh failure containing '$expected'; status=$status output=$output" >&2
    exit 1
  fi
}
setup_repo
printf 'new（new）\nstable\n' > internal/i18n/locales/ja/game.json
expect_failure 'FAIL: repeated Japanese term'
setup_repo
printf 'stable\n' > internal/i18n/locales/ja/game.json
expect_failure 'FAIL: line count changed'
setup_repo
printf 'new term\nstable\n' > internal/i18n/locales/ja/game.json
bash ./after.sh old new internal/i18n/locales/ja/game.json

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
