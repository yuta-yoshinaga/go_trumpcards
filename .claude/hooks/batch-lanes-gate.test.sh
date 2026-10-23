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
  mkdir -p "$repo/frontend/src/pages" "$repo/frontend/e2e" "$base/scripts" "$state"
  for spec in desktop-board-visibility desktop-hand-visibility game-message-visibility mobile-viewport text-contrast route-sweep; do
    touch "$repo/frontend/e2e/$spec.spec.ts"
  done
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
  for script in common.sh allowlist.sh rc.sh rcg.sh treehash.sh e2egate.sh; do cp "$SCRIPTS/$script" "$base/scripts/"; done
  printf '#!/usr/bin/env bash\nexit 0\n' > "$base/scripts/e2e.sh"
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
run_gap_gate preexisting 'if (ready) {' 'PREEXISTING src/pages/sample.tsx:1:branch [0] | if (ready) {' 'NOTE PATCH_BRANCH_GAP_PREEXISTING' yes
run_gap_gate new_text 'if (ready) {' 'src/pages/sample.tsx:1:branch [0] | if (fresh) {' 'WARN PATCH_BRANCH_GAP' no
run_gap_gate quoted_warn 'label="WARN FAIL"' 'PREEXISTING src/pages/sample.tsx:1:branch [0] | label="WARN FAIL"' 'NOTE PATCH_BRANCH_GAP_PREEXISTING' yes
run_gap_gate short_text 'x ? a : b' 'src/pages/sample.tsx:1:branch [0] | x ?' 'WARN PATCH_BRANCH_GAP' no
run_gap_gate new_file 'if (ready) {' 'missing.tsx:1:branch [0] | if (fresh) {' 'WARN PATCH_BRANCH_GAP' no

# A gate stamp must be rejected when E2E changes the worktree after the receipt.
race="$tmp/e2e-race"; mkdir -p "$race/repo" "$race/wts" "$race/scripts" "$race/state"
git init -q "$race/repo"; git -C "$race/repo" config user.email test@example.com; git -C "$race/repo" config user.name test
printf 'base\n' > "$race/repo/file"; git -C "$race/repo" add file; git -C "$race/repo" commit -qm base
git -C "$race/repo" worktree add -q -b lane "$race/wts/wt-ib1" HEAD
for script in common.sh allowlist.sh treehash.sh rcg.sh; do cp "$SCRIPTS/$script" "$race/scripts/"; done
cat > "$race/scripts/rc.sh" <<'STUB'
#!/usr/bin/env bash
echo CHECK=ok > "$BATCH_STATE/rc$1.txt"
echo TYPE=ok >> "$BATCH_STATE/rc$1.txt"
STUB
cat > "$race/scripts/e2egate.sh" <<'STUB'
#!/usr/bin/env bash
printf 'changed during E2E\n' >> "$BATCH_WT_ROOT/wt-ib$1/file"
STUB
chmod +x "$race/scripts/"*.sh
output="$(BATCH_REPO="$race/repo" BATCH_WT_ROOT="$race/wts" BATCH_STATE="$race/state" bash "$race/scripts/rcg.sh" 1 2>&1)" && fail "rcg accepted a worktree change during E2E"
grep -q 'GATE: worktree changed during e2e' <<< "$output" || fail "rcg omitted worktree race error: $output"
[[ ! -e "$race/state/gate1.ok" ]] || fail "rcg stamped a worktree changed during E2E"

# A queued receipt holding the shared lock must not delay post-E2E hash/stamp work.
queued="$tmp/e2e-queued-lock"; mkdir -p "$queued/repo" "$queued/wts" "$queued/scripts" "$queued/state"
git init -q "$queued/repo"; git -C "$queued/repo" config user.email test@example.com; git -C "$queued/repo" config user.name test
printf 'base\n' > "$queued/repo/file"; git -C "$queued/repo" add file; git -C "$queued/repo" commit -qm base
git -C "$queued/repo" worktree add -q -b lane "$queued/wts/wt-ib1" HEAD
for script in common.sh allowlist.sh treehash.sh rcg.sh; do cp "$SCRIPTS/$script" "$queued/scripts/"; done
cat > "$queued/scripts/rc.sh" <<'STUB'
#!/usr/bin/env bash
echo CHECK=ok > "$BATCH_STATE/rc$1.txt"
echo TYPE=ok >> "$BATCH_STATE/rc$1.txt"
STUB
cat > "$queued/scripts/e2egate.sh" <<'STUB'
#!/usr/bin/env bash
(flock -x "$BATCH_STATE/receipt.lock" sleep 15) >/dev/null 2>&1 </dev/null &
exit 0
STUB
chmod +x "$queued/scripts/"*.sh
start_time=$SECONDS
output="$(BATCH_REPO="$queued/repo" BATCH_WT_ROOT="$queued/wts" BATCH_STATE="$queued/state" timeout 5 bash "$queued/scripts/rcg.sh" 1 2>&1)" || fail "rcg waited on queued receipt lock: $output"
[[ $((SECONDS - start_time)) -lt 5 ]] || fail "rcg exceeded five seconds with queued receipt lock"
[[ -s "$queued/state/gate1.ok" ]] || fail "rcg did not stamp gate success while queued receipt held lock"

# Lint output without an explicit success receipt must not pass the gate.
lint_gate() {
  local name="$1" lint_lines="$2" want_pass="$3" expected_gate="${4:-GATE: lint unverified}"
  local base="$tmp/$name" repo="$tmp/$name/repo" wts="$tmp/$name/wts" state="$tmp/$name/state"
  mkdir -p "$repo" "$wts" "$base/scripts" "$state"
  git init -q "$repo"; git -C "$repo" config user.email test@example.com; git -C "$repo" config user.name test
  printf 'base\n' > "$repo/file"; git -C "$repo" add file; git -C "$repo" commit -qm base
  git -C "$repo" worktree add -q -b lane "$wts/wt-ib1" HEAD
  for script in common.sh allowlist.sh treehash.sh rcg.sh; do cp "$SCRIPTS/$script" "$base/scripts/"; done
  printf '#!/usr/bin/env bash\nexit 0\n' > "$base/scripts/e2egate.sh"
  cat > "$base/scripts/rc.sh" <<STUB
#!/usr/bin/env bash
{
  echo CHECK=ok
  echo TYPE=ok
$lint_lines
  echo E2E_GATE=0
} > "\$BATCH_STATE/rc\$1.txt"
STUB
  chmod +x "$base/scripts/"*.sh
  local output rc=0
  output="$(BATCH_REPO="$repo" BATCH_WT_ROOT="$wts" BATCH_STATE="$state" E2E_GATE=0 bash "$base/scripts/rcg.sh" 1 2>&1)" && rc=0 || rc=$?
  [[ "$rc" = 2 && "$output" == *'E2E_GATE=0 requires E2E_SKIP_REASON'* ]] || fail 'rcg did not reject an E2E bypass without a reason'
  rc=0
  output="$(BATCH_REPO="$repo" BATCH_WT_ROOT="$wts" BATCH_STATE="$state" E2E_GATE=0 E2E_SKIP_REASON='unit test bypass' bash "$base/scripts/rcg.sh" 1 2>&1)" || rc=$?
  if [[ "$want_pass" == yes ]]; then
    [[ $rc -eq 0 ]] || fail "$name gate should pass: $output"
    grep -q '^GATE: ok$' <<< "$output" || fail "$name omitted success: $output"
    [[ -s "$state/gate1.ok" ]] || fail "$name did not stamp gate ok"
    if [[ "$name" == lint_verified ]]; then
      output="$(BATCH_REPO="$repo" BATCH_WT_ROOT="$wts" BATCH_STATE="$state" bash "$base/scripts/rcg.sh" 1 2>&1)" || fail "normal gate rerun failed: $output"
      ! grep -q '^NOTE E2E_GATE skipped' "$state/rc1.txt" || fail 'normal gate rerun retained stale E2E skip receipt'
      unset BATCH_E2E_SKIP_REASON
      BATCH_E2E_SKIP_REASON="$(sed -n 's/^NOTE E2E_GATE skipped: //p' "$state/rc1.txt" | tail -1)"
      source "$SCRIPTS/prbody.sh"
      body=$(ship_pr_body 42 '- rerun gate' '' '')
      [[ "$body" != *'E2E skipped:'* ]] || fail "ship body retained stale E2E skip reason: $body"
    fi
  else
    [[ $rc -ne 0 ]] || fail "$name gate should fail: $output"
    grep -q "^$expected_gate$" <<< "$output" || fail "$name omitted lint error: $output"
    [[ ! -e "$state/gate1.ok" ]] || fail "$name stamped unverified lint"
  fi
}
lint_gate lint_unverified '  echo "LINT: Error: parallel golangci-lint is running"' no
lint_gate lint_verified $'  echo "LINT: 0 issues."\n  echo LINT=ok' yes
lint_gate lint_failed $'  echo "LINT: 2 issues."\n  echo LINT=FAIL' no 'GATE: lint failed'

for script in "$SCRIPTS/receipt.sh" "$SCRIPTS/rcg.sh"; do bash -n "$script" || fail "bash -n $script"; done
grep -q -- '--allow-parallel-runners' "$SCRIPTS/receipt.sh" || fail 'receipt.sh omits parallel-runner flag'
grep -q '^  if \[\[ \$lint_rc -eq 0 \]\]; then echo LINT=ok;' "$SCRIPTS/receipt.sh" || fail 'receipt.sh omits LINT=ok receipt'
echo "batch-lanes gate tests passed"
