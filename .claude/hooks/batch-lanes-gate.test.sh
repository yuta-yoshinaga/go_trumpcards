#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPTS="$ROOT/.claude/skills/batch-lanes/scripts"
fail() { echo "FAIL: $*" >&2; exit 1; }
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
repo="$tmp/repo"
git init -q "$repo"
git -C "$repo" config user.email test@example.com
git -C "$repo" config user.name test
mkdir -p "$repo/internal" "$repo/frontend"
echo 'package internal' > "$repo/internal/base.go"
echo 'export {}' > "$repo/frontend/base.ts"
git -C "$repo" add . && git -C "$repo" commit -qm base
cat > "$repo/internal/new.go" <<'GO'
package internal

func NewUnusedThing() {}

// DocUnusedThing is a doc-commented function.
func DocUnusedThing() {}
GO
cat > "$repo/frontend/new.ts" <<'TS'
export function unusedNewThing() {}

/**
 * unusedJSDocThing has a JSDoc comment.
 */
export function unusedJSDocThing() {}
TS
out="$(bash "$SCRIPTS/unusedsym.sh" "$repo")"
grep -q '^WARN UNUSED_NEW_SYMBOL NewUnusedThing internal/new.go$' <<< "$out" || fail "missing unused Go warning: $out"
grep -q '^WARN UNUSED_NEW_SYMBOL DocUnusedThing internal/new.go$' <<< "$out" || fail "missing doc-commented unused Go warning: $out"
grep -q '^WARN UNUSED_NEW_SYMBOL unusedNewThing frontend/new.ts$' <<< "$out" || fail "missing unused TS warning: $out"
grep -q '^WARN UNUSED_NEW_SYMBOL unusedJSDocThing frontend/new.ts$' <<< "$out" || fail "missing JSDoc unused TS warning: $out"
cat > "$repo/internal/use.go" <<'GO'
package internal
func UseIt() {
	NewUnusedThing()
	DocUnusedThing()
}
GO
cat > "$repo/frontend/use.ts" <<'TS'
import { unusedNewThing, unusedJSDocThing } from './new'
unusedNewThing()
unusedJSDocThing()
TS
out="$(bash "$SCRIPTS/unusedsym.sh" "$repo")"
! grep -q 'WARN UNUSED_NEW_SYMBOL NewUnusedThing\|WARN UNUSED_NEW_SYMBOL DocUnusedThing\|WARN UNUSED_NEW_SYMBOL unusedNewThing\|WARN UNUSED_NEW_SYMBOL unusedJSDocThing' <<< "$out" || fail "reported referenced symbol: $out"
out="$(ALLOW_UNUSED='NewUnusedThing DocUnusedThing unusedNewThing unusedJSDocThing UseIt' bash "$SCRIPTS/unusedsym.sh" "$repo")"
! grep -q 'WARN UNUSED_NEW_SYMBOL' <<< "$out" || fail "ALLOW_UNUSED did not suppress warnings: $out"

BATCH_TESTING=1 source "$SCRIPTS/workerbuild.sh"
tags="$(worker_tags_from_line '//go:build !js || !wasm || extra2')"
[[ "$tags" == extra2 ]] || fail "worker build tag parsing was unexpected: $tags"
! worker_tags_from_line '//go:build js && wasm' | grep -q . || fail "js/wasm-only constraint produced a worker tag"
for check in workerbuild.sh unusedsym.sh; do bash -n "$SCRIPTS/$check" || fail "bash -n $check"; done
grep -q 'workerbuild.sh' "$SCRIPTS/rc.sh" || fail "rc.sh does not call workerbuild.sh"
grep -q 'unusedsym.sh' "$SCRIPTS/rc.sh" || fail "rc.sh does not call unusedsym.sh"
grep -q 'patch-branch-gaps.sh' "$SCRIPTS/rc.sh" || fail "rc.sh does not call patch-branch-gaps.sh"

# The gate must distinguish unchanged develop branches from newly introduced gaps.
run_gap_gate() {
  local name="$1" develop_text="$2" gap="$3" classification="$4" want_pass="$5"
  local base="$tmp/$name" repo="$tmp/$name/repo" wtroot="$tmp/$name/wts" state="$tmp/$name/state"
  mkdir -p "$repo/frontend/src/pages" "$base/scripts" "$state"
  git init -q "$repo"
  git -C "$repo" config user.email test@example.com
  git -C "$repo" config user.name test
  printf '%s\n' "$develop_text" > "$repo/frontend/src/pages/sample.tsx"
  git -C "$repo" add . && git -C "$repo" commit -qm develop
  git -C "$repo" branch -M develop
  git -C "$repo" update-ref refs/remotes/origin/develop HEAD
  mkdir -p "$wtroot"
  git -C "$repo" clone -q . "$wtroot/wt-ib1"
  mkdir -p "$wtroot/wt-ib1/.claude/skills/patch-branch-gaps/scripts"
  printf '\n// change\n' >> "$wtroot/wt-ib1/frontend/src/pages/sample.tsx"
  printf '#!/usr/bin/env bash\nprintf "%%s\\n" "$GAP_OUTPUT"\nexit 1\n' > "$wtroot/wt-ib1/.claude/skills/patch-branch-gaps/scripts/patch-branch-gaps.sh"
  for script in common.sh rc.sh rcg.sh treehash.sh e2egate.sh; do cp "$SCRIPTS/$script" "$base/scripts/"; done
  for stub in workerbuild unusedsym repeattests biomeignore; do printf '#!/usr/bin/env bash\nexit 0\n' > "$base/scripts/$stub.sh"; done
  printf '#!/usr/bin/env bash\necho CHECK=ok\necho TYPE=ok\n' > "$base/scripts/receipt.sh"
  chmod +x "$base/scripts/"*.sh
  local output rc=0
  output="$(BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$state" GAP_OUTPUT="$gap" ALLOW_WORKER_BUILD=1 REPEAT_TESTS_COUNT=0 bash "$base/scripts/rcg.sh" 1 2>&1)" || rc=$?
  grep -q "$classification" "$state/rc1.txt" || fail "$name missing $classification: $output"
  grep -q "$classification" <<< "$output" || fail "$name summary omitted $classification: $output"
  if [[ "$want_pass" == yes ]]; then [[ $rc -eq 0 ]] || fail "$name gate should pass: $output"
  else [[ $rc -ne 0 ]] || fail "$name gate should fail: $output"; fi
}
run_gap_gate preexisting 'if (ready) {' 'src/pages/sample.tsx:1:branch [0] | if (ready) {' 'NOTE PATCH_BRANCH_GAP_PREEXISTING' yes
run_gap_gate new_text 'if (ready) {' 'src/pages/sample.tsx:1:branch [0] | if (fresh) {' 'WARN PATCH_BRANCH_GAP' no
run_gap_gate quoted_warn 'label="WARN FAIL"' 'src/pages/sample.tsx:1:branch [0] | label="WARN FAIL"' 'NOTE PATCH_BRANCH_GAP_PREEXISTING' yes
run_gap_gate short_text 'x ? a : b' 'src/pages/sample.tsx:1:branch [0] | x ?' 'WARN PATCH_BRANCH_GAP' no
run_gap_gate new_file 'if (ready) {' 'missing.tsx:1:branch [0] | if (fresh) {' 'WARN PATCH_BRANCH_GAP' no
echo "batch-lanes gate tests passed"
