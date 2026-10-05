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
echo "batch-lanes gate tests passed"
