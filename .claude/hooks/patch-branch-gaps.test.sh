#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/.claude/skills/patch-branch-gaps/scripts/patch-branch-gaps.sh"
fail() { echo "FAIL: $*" >&2; exit 1; }
bash -n "$SCRIPT" || fail "patch-branch-gaps.sh syntax"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin"
cat > "$tmp/bin/bunx" <<'STUB'
#!/usr/bin/env bash
for arg in "$@"; do
  case "$arg" in
    --coverage.include=*) source_path="${arg#*=}" ;;
    --coverage.reportsDirectory=*) coverage_dir="${arg#*=}" ;;
  esac
done
mkdir -p "$coverage_dir"
printf '{"%s":{"branchMap":{"0":{"type":"cond-expr","loc":{"start":{"line":1}}}},"b":{"0":[0,1]}}}\n' "/tmp/$source_path" > "$coverage_dir/coverage-final.json"
STUB
chmod +x "$tmp/bin/bunx"

run_case() {
  local name="$1" develop_line="$2" changed_line="$3" expected="$4" wanted_rc="$5"
  local repo="$tmp/$name"
  mkdir -p "$repo/frontend/src/utils"
  git init -q "$repo"
  git -C "$repo" config user.email test@example.com
  git -C "$repo" config user.name test
  printf '%s\n' "$develop_line" > "$repo/frontend/src/utils/sample.ts"
  printf 'test placeholder\n' > "$repo/frontend/src/utils/sample.test.ts"
  git -C "$repo" add . && git -C "$repo" commit -qm develop
  git -C "$repo" update-ref refs/remotes/origin/develop HEAD
  printf '%s\n' "$changed_line" > "$repo/frontend/src/utils/sample.ts"
  local output rc=0
  output="$(cd "$repo" && PATH="$tmp/bin:$PATH" bash "$SCRIPT" 2>&1)" || rc=$?
  [[ $rc -eq $wanted_rc ]] || fail "$name exit $rc, wanted $wanted_rc: $output"
  grep -Fq "$expected" <<< "$output" || fail "$name missing '$expected': $output"
  if [[ "$expected" == PREEXISTING* ]]; then
    grep -Fq 'summary: 0 new, 1 pre-existing' <<< "$output" || fail "$name wrong existing count: $output"
  else
    grep -Fq 'summary: 1 new, 0 pre-existing' <<< "$output" || fail "$name wrong new count: $output"
  fi
}

# Trimmed exact text and a branch shifted up from line 2 both match develop.
run_case exact 'if (ready) {' '  if (ready) {' 'PREEXISTING src/utils/sample.ts:1:' 0
run_case shifted $'const x = 0\nif (ready) { // old location' 'if (ready) {' 'PREEXISTING src/utils/sample.ts:1:' 0
run_case fresh 'if (ready) {' 'if (fresh) {' 'src/utils/sample.ts:1: cond-expr [0, 1] | if (fresh) {' 1
# 7 characters and a prefix of the develop line: only the 8-character minimum keeps it new.
run_case short 'x ? a : b' 'x ? a' 'src/utils/sample.ts:1: cond-expr [0, 1] | x ? a' 1
echo "patch-branch-gaps tests passed"
