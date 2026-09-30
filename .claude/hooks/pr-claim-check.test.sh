#!/usr/bin/env bash
set -uo pipefail
HOOK="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/pr-claim-check.sh"
TMP=$(mktemp -d); trap 'rm -rf "$TMP"' EXIT
cd "$TMP" || exit 1
git init -q .; git config user.email t@example.com; git config user.name test
mkdir -p locales; printf '{"trickSummary":"このトリック: {{points}}点"}\n' > locales/ja.json
git add .; git commit -qm init; git branch -M develop; git branch -f origin/develop HEAD 2>/dev/null || true
fail=0
run() {
  local expect=$1 desc=$2 command=$3 out denied=pass
  out=$(jq -nc --arg command "$command" '{tool_input:{command:$command}}' | bash "$HOOK")
  printf '%s' "$out" | grep -q '"permissionDecision":"deny"' && denied=deny
  if [ "$denied" = "$expect" ]; then echo "  ok    $desc ($expect)"; else echo "  FAIL  $desc expected $expect: $out"; fail=1; fi
}
echo 'pr-claim-check:'
run pass 'existing quote' 'gh pr comment 1 --body "`このトリック`"'
run pass 'digit split existing quote' 'gh pr create --body "`このトリック: 5.5点`"'
run deny 'digit split missing quote' 'gh pr create --body "`このトリックの合計: 5点`"'
run deny 'missing quote' 'gh pr create --body "`上段 ハイカード に勝ち`"'
run pass 'skip marker' 'gh pr create --body "`上段 ハイカード に勝ち` claim-check: skip"'
run pass 'ASCII quote' 'gh pr create --body "`someIdentifier`"'
printf '`存在しない引用`\n' > "$TMP/body.md"
run deny 'body file missing quote' 'gh pr create --body-file body.md'
run pass 'non-gh command' 'echo "`存在しない引用`"'
run pass 'gh pr view' 'gh pr view --body "`存在しない引用`"'
[ "$fail" -eq 0 ] || exit 1
echo 'pr-claim-check: all cases passed'
