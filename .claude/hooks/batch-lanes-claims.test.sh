#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPTS="$ROOT/.claude/skills/batch-lanes/scripts"
fail() { echo "FAIL: $*" >&2; exit 1; }
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
repo="$tmp/repo"
git init -q "$repo"
mkdir -p "$repo/pkg/deep"
mkdir -p "$repo/frontend"
mkdir -p "$repo/node_modules/fake"
printf 'node_modules/\n' >"$repo/.gitignore"
cat >"$repo/pkg/sample_test.go" <<'EOF'
package pkg
func TestExistingCase(t *testing.T) {}
func TestCases(t *testing.T) { t.Run("nested case", func(t *testing.T) {}) }
EOF
cat >"$repo/frontend/sample.test.ts" <<'EOF'
it('existing browser case', () => {})
EOF
printf 'GetTrickWinnerIdx and pointLimit\n' >"$repo/pkg/source.go"
printf 'one\ntwo\nthree\nfour\n' >"$repo/pkg/deep/Foo.go"
git -C "$repo" add .
git -C "$repo" -c user.name=test -c user.email=test@example.com commit -qm fixture
printf 'it("node-only case", () => {})\n' >"$repo/node_modules/fake/only.test.ts"
bash "$SCRIPTS/claimcheck.sh" "$repo" 'Uses `GetTrickWinnerIdx`, `state.config.pointLimit`, func TestExistingCase(t *testing.T), t.Run("nested case"), it('\''existing browser case'\''), and pkg/sample_test.go:2' || fail "valid claims rejected"
bash "$SCRIPTS/claimcheck.sh" "$repo" 'Foo.go:3' || fail "basename file:line claim did not match a deep file"
set +e
out=$(bash "$SCRIPTS/claimcheck.sh" "$repo" 'Foo.go:5' 2>&1)
rc=$?
set -e
[ "$rc" -eq 1 ] || fail "short basename file:line returned $rc"
grep -q '^UNVERIFIED file:line Foo.go:5$' <<<"$out" || fail "short basename file:line not reported: $out"
set +e
out=$(bash "$SCRIPTS/claimcheck.sh" "$repo" 'Missing `NoSuchExport`, func TestMissingCase(t *testing.T), and pkg/missing.go:8' 2>&1)
rc=$?
set -e
[ "$rc" -eq 1 ] || fail "missing claims returned $rc"
grep -q '^UNVERIFIED identifier NoSuchExport$' <<<"$out" || fail "missing identifier not reported: $out"
grep -q '^UNVERIFIED test TestMissingCase$' <<<"$out" || fail "missing test not reported: $out"
grep -q '^UNVERIFIED file:line pkg/missing.go:8$' <<<"$out" || fail "missing path not reported: $out"
set +e
out=$(bash "$SCRIPTS/claimcheck.sh" "$repo" 'it("node-only case")' 2>&1)
rc=$?
set -e
[ "$rc" -eq 1 ] || fail "node_modules-only test claim returned $rc"
grep -q '^UNVERIFIED test node-only case$' <<<"$out" || fail "node_modules-only test was accepted: $out"
bash "$SCRIPTS/claimcheck.sh" "$repo" 'claim-check: skip `NoSuchExport`' || fail "skip marker rejected"
bash "$SCRIPTS/claimcheck.sh" "$repo" '日本語の説明 with general English words and no quoted code' || fail "ordinary prose rejected"
grep -q 'claimcheck.sh' "$SCRIPTS/ship.sh" || fail "ship.sh does not call claimcheck"
grep -q 'claimcheck.sh' "$SCRIPTS/fixpush.sh" || fail "fixpush.sh does not call claimcheck"
grep -q 'serverHint' "$SCRIPTS/mkprompt.sh" || fail "mkprompt.sh missing serverHint"
grep -q 'design.*BACK' "$SCRIPTS/mkprompt.sh" || fail "mkprompt.sh missing design BACK rule"
grep -q 'go build -tags' "$SCRIPTS/mkprompt.sh" || fail "mkprompt.sh missing worker build rule"
if grep -q '\\"' "$SCRIPTS/mkprompt.sh"; then fail 'mkprompt.sh contains escaped quotes in an unquoted heredoc'; fi
heredoc="$tmp/mkprompt-heredoc.sh"
sed -n '/^cat <<P$/,/^P$/p' "$SCRIPTS/mkprompt.sh" >"$heredoc"
prompt=$(bash -c 'n=1; body=x; source /dev/stdin' <"$heredoc") || fail "mkprompt heredoc did not expand"
grep -q 'design "BACK"' <<<"$prompt" || fail "expanded prompt lost design BACK quotes"
grep -q '`VERDICT: NOT_PLANNED`' <<<"$prompt" || fail "expanded prompt lost backtick-delimited text"
grep -q '\$ref' <<<"$prompt" || fail "expanded prompt lost literal dollar text"
if grep -q '\\"' <<<"$prompt"; then fail "expanded prompt contains escaped quotes"; fi
for script in ship.sh fixpush.sh; do
  if [[ "$script" == ship.sh ]]; then wt_path='\$WT_ROOT/wt-ib\$s'; else wt_path='\$WT_ROOT/wt-ib\$1'; fi
  cd_line=$(grep -n "^cd \"$wt_path\"" "$SCRIPTS/$script" | cut -d: -f1)
  check_line=$(grep -n 'claimcheck.sh.*"\$PWD"' "$SCRIPTS/$script" | cut -d: -f1)
  refused_line=$(grep -n 'REFUSED: unverified claims' "$SCRIPTS/$script" | cut -d: -f1)
  [[ -n "$cd_line" && -n "$check_line" && -n "$refused_line" && "$cd_line" -lt "$check_line" && "$check_line" -lt "$refused_line" ]] || fail "$script claimcheck order/worktree path is wrong"
done
echo "batch-lanes claims tests passed"
