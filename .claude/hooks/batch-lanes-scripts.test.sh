#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPTS="$ROOT/.claude/skills/batch-lanes/scripts"
fail() { echo "FAIL: $*" >&2; exit 1; }

for script in "$SCRIPTS"/*.sh; do bash -n "$script" || fail "bash -n $script"; done
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
# Regression checks for review findings in the batch lane helpers.
for script in ship.sh fixpush.sh; do
  grep -q 'PIPESTATUS' "$SCRIPTS/$script" || fail "$script does not preserve push status"
  grep -q 'push_status == 0' "$SCRIPTS/$script" || fail "$script does not stop after push failure"
  grep -q 'git diff --name-only -z HEAD' "$SCRIPTS/$script" || fail "$script does not collect paths with NUL delimiters"
done
! grep -R -nE 'git status --porcelain \| awk' "$SCRIPTS" || fail "status paths are parsed as whitespace fields"
if grep -R -n '/tmp/' "$SCRIPTS" >"$tmp/tmp-paths.out" 2>&1; then
  fail "batch scripts contain hardcoded /tmp paths: $(cat "$tmp/tmp-paths.out")"
fi
! grep -q 'Co-Authored-By: Claude Opus 5.5' "$SCRIPTS/ship.sh" || fail "ship.sh hardcodes co-author"
! grep -q 'gh pr view .*codecov' "$SCRIPTS/land.sh" || fail "land.sh retains result-free codecov pipeline"
grep -q 'CLAUDE_CO_AUTHOR' "$SCRIPTS/ship.sh" || fail "ship.sh does not support CLAUDE_CO_AUTHOR"
grep -q 'BATCH_DELEGATED_TO' "$SCRIPTS/ship.sh" || fail "ship.sh does not support BATCH_DELEGATED_TO"
grep -q 'Delegated-To:.*BATCH_DELEGATED_TO:-codex (dele -k edit)' "$SCRIPTS/fixpush.sh" || fail "fixpush.sh does not retain the default delegated-to metadata"
grep -q 'gh pr merge .*>/dev/null' "$SCRIPTS/land.sh" || fail "land.sh does not suppress merge stdout"
grep -q 'Failed to merge PR' "$SCRIPTS/land.sh" || fail "land.sh does not report merge failures"
grep -q 'go_trumpcards 固有の規約' "$SCRIPTS/receipt.sh" || fail "receipt.sh does not explain project-specific conventions"
grep -qF '[ -f "$t.test.ts" ] || [ -f "$t.test.tsx" ]' "$SCRIPTS/receipt.sh" || fail "receipt.sh MISSING_TEST must accept a .test.tsx file for a .ts source (hooks are often tested with providers in .tsx)"
grep -qP '(?<!\\)`' "$SCRIPTS/mkprompt.sh" && fail "mkprompt.sh has an unescaped backtick: its heredoc is unquoted, so bash runs the quoted text as a command and drops it from every lane prompt"
sed -n '/^cat <<P$/,/^P$/p' "$SCRIPTS/mkprompt.sh" | grep -qP '(?<!\\)\$(?!n\b|body\b)' && fail "mkprompt.sh has a bare dollar in its unquoted heredoc; bash expands it to empty and removes a word from the instructions"
grep -q 'dele' "$ROOT/.claude/skills/batch-lanes/SKILL.md" || fail "SKILL.md does not document dele"
set +e
grep -rnE '/tmp/claude-1000|/home/yuta|session_0' "$SCRIPTS" >"$tmp/paths.out" 2>&1
grep_rc=$?
set -e
if [ "$grep_rc" -eq 0 ]; then fail "machine/session path is hardcoded"; fi
[ "$grep_rc" -eq 1 ] || fail "grep failed while checking machine/session paths: $(cat "$tmp/paths.out")"
export BATCH_STATE="$tmp/state"
cat >"$tmp/dele" <<'STUB'
#!/usr/bin/env bash
touch "$DELE_CALLED"
STUB
chmod +x "$tmp/dele"
if PATH="$tmp:$PATH" DELE_CALLED="$tmp/called" \
  bash "$SCRIPTS/go.sh" 1 123 "$tmp/missing.md" >/dev/null 2>&1; then
  fail "go.sh accepted a missing prompt"
fi
[ ! -e "$tmp/called" ] || fail "go.sh invoked dele without a prompt"
mkdir -p "$tmp/wt-ib1"
printf 'delegate this issue\n' > "$tmp/prompt.md"
PATH="$tmp:$PATH" DELE_CALLED="$tmp/called" BATCH_WT_ROOT="$tmp" \
  bash "$SCRIPTS/go.sh" 1 123 "$tmp/prompt.md" >/dev/null 2>&1 || fail "go.sh rejected a non-empty prompt"
[ -e "$tmp/called" ] || fail "go.sh did not invoke dele for a valid prompt"

if BATCH_REPO="$ROOT" bash "$SCRIPTS/sweep.sh" >"$tmp/sweep.out" 2>&1; then
  fail "sweep.sh accepted an unset BATCH_BRANCH_RE"
fi
grep -qi 'BATCH_BRANCH_RE' "$tmp/sweep.out" || fail "sweep.sh did not explain BATCH_BRANCH_RE"
cat >"$tmp/gh" <<'STUB'
#!/usr/bin/env bash
case "$*" in
  *"pr list"*"number,headRefName"*) printf '%s\n' '[{"number":123,"headRefName":"feat/123-good"},{"number":456,"headRefName":"chore/456-skip"}]' ;;
  *"pr list"*"headRefName"*)
    if [ "${GH_EMPTY_LIST:-}" = 1 ]; then echo '[]'; exit 0; fi
    count_file="$GH_CALLS/count"
    count=0; [ ! -f "$count_file" ] || count=$(cat "$count_file")
    count=$((count + 1)); printf '%s\n' "$count" > "$count_file"
    if [ "$count" -eq 1 ]; then printf '%s\n' '[{"headRefName":"feat/123-good"},{"headRefName":"chore/456-skip"}]'; else printf '%s\n' '[]'; fi ;;
  *"pr view 123"*"mergeable"*) echo '"MERGEABLE"' ;;
  *"pr checks 123"*) echo '[{"bucket":"pending","name":"CI"}]' ;;
  *"pr view 123"*"comments"*"length"*) echo '1' ;;
  *"pr view 123"*"comments"*"body"*) printf 'review finished\n---\n' ;;
  *) echo "unexpected gh invocation: $*" >&2; exit 2 ;;
esac
STUB
chmod +x "$tmp/gh"
mkdir -p "$tmp/empty-state"
PATH="$tmp:$PATH" GH_CALLS="$tmp" BATCH_REPO="$ROOT" BATCH_STATE="$tmp/empty-state" BATCH_BRANCH_RE='^feat/[0-9]+-' \
  bash "$SCRIPTS/triage.sh" >"$tmp/empty-triage.out" 2>"$tmp/empty-triage.err" || fail "triage.sh failed with empty state"
if grep -q 'No such file' "$tmp/empty-triage.err"; then fail "triage.sh reported a missing state file: $(cat "$tmp/empty-triage.err")"; fi
mkdir -p "$tmp/state"
touch "$tmp/state/reviewed.txt"
PATH="$tmp:$PATH" GH_CALLS="$tmp" BATCH_REPO="$ROOT" BATCH_STATE="$tmp/state" BATCH_BRANCH_RE='^feat/[0-9]+-' \
  bash "$SCRIPTS/sweep.sh" >"$tmp/sweep.out" || fail "sweep.sh rejected a configured branch expression"
grep -q '^REVIEW-READY 123$' "$tmp/sweep.out" || fail "sweep.sh did not filter the matching PR: $(cat "$tmp/sweep.out")"
grep -q '^UNSWEPT 456 chore/456-skip$' "$tmp/sweep.out" || fail "sweep.sh did not report the nonmatching PR as UNSWEPT"
PATH="$tmp:$PATH" GH_CALLS="$tmp" BATCH_REPO="$ROOT" BATCH_STATE="$tmp/state" BATCH_BRANCH_RE='^feat/[0-9]+-' \
  bash "$SCRIPTS/triage.sh" >"$tmp/triage.out" || fail "triage.sh failed"
grep -q '^OK 123$' "$tmp/triage.out" || fail "triage.sh did not filter the matching PR: $(cat "$tmp/triage.out")"
PATH="$tmp:$PATH" GH_EMPTY_LIST=1 GH_CALLS="$tmp" BATCH_REPO="$ROOT" BATCH_STATE="$tmp/state" BATCH_BRANCH_RE='^feat/[0-9]+-' \
  bash "$SCRIPTS/drain.sh" >"$tmp/drain.out" || fail "drain.sh failed"
grep -q 'ALL MERGED' "$tmp/drain.out" || fail "drain.sh did not evaluate the filtered PR list: $(cat "$tmp/drain.out")"

# fx.sh: a clean checkout of the target branch in another worktree is detached;
# dirty work is reported and left untouched.
git init -q "$tmp/repo"
git -C "$tmp/repo" config user.email test@example.com
git -C "$tmp/repo" config user.name test
mkdir -p "$tmp/repo/frontend/node_modules"
echo fixture > "$tmp/repo/frontend/fixture"
echo original > "$tmp/repo/file"
git -C "$tmp/repo" add . && git -C "$tmp/repo" commit -qm initial
git -C "$tmp/repo" branch target
git -C "$tmp/repo" branch develop
git init -q --bare "$tmp/remote.git"
git -C "$tmp/repo" remote add origin "$tmp/remote.git"
git -C "$tmp/repo" push -q origin target develop
git -C "$tmp/repo" fetch -q origin
git -C "$tmp/repo" worktree add -q "$tmp/other" target
git -C "$tmp/repo" worktree add -q --detach "$tmp/wt-ibfix" origin/develop
PATH="$PATH" BATCH_REPO="$tmp/repo" BATCH_WT_ROOT="$tmp" \
  bash "$SCRIPTS/fx.sh" fix target >"$tmp/fx.out" || fail "fx.sh failed to detach a clean worktree"
git -C "$tmp/other" symbolic-ref -q HEAD && fail "fx.sh left the target branch checked out elsewhere"
git -C "$tmp/other" status --porcelain | grep -q . && fail "fx.sh dirtied the detached worktree"
git -C "$tmp/wt-ibfix" checkout -q --detach origin/develop
git -C "$tmp/other" checkout -q target
echo changed > "$tmp/other/file"
if PATH="$PATH" BATCH_REPO="$tmp/repo" BATCH_WT_ROOT="$tmp" \
  bash "$SCRIPTS/fx.sh" fix target >"$tmp/fx-dirty.out" 2>&1; then fail "fx.sh accepted a dirty worktree on target"; fi
grep -q "$tmp/other" "$tmp/fx-dirty.out" || fail "fx.sh did not report the dirty worktree path"

# e2echeck should ignore removed locale text found only on comment lines.
for case_name in comment locator; do
  repo="$tmp/e2echeck-$case_name"
  git init -q "$repo"
  git -C "$repo" config user.email test@example.com
  git -C "$repo" config user.name test
  mkdir -p "$repo/frontend/src/i18n/locales/ja" "$repo/frontend/e2e"
  printf '{\n  "message": "削除対象の文言です"\n}\n' > "$repo/frontend/src/i18n/locales/ja/bura.json"
  if [ "$case_name" = comment ]; then
    printf '// 削除対象の文言です\n' > "$repo/frontend/e2e/bura.spec.ts"
  else
    printf 'getByText("削除対象の文言です");\n' > "$repo/frontend/e2e/bura.spec.ts"
  fi
  git -C "$repo" add .
  printf '{\n}\n' > "$repo/frontend/src/i18n/locales/ja/bura.json"
  output="$(python3 "$SCRIPTS/e2echeck.py" "$repo")" || fail "e2echeck failed for $case_name case"
  if [ "$case_name" = comment ]; then
    [ -z "$output" ] || fail "e2echeck warned for comment-only hit: $output"
  else
    [[ "$output" == *"E2ECHECK WARN"* ]] || fail "e2echeck missed locator hit"
  fi
done

BATCH_TESTING=1 source "$SCRIPTS/receipt.sh"
BATCH_STATE="$tmp/testing-state" BATCH_TESTING=1 bash "$SCRIPTS/receipt.sh" 1 || fail "receipt.sh testing mode failed when executed directly"
[ "$(page_test_for frontend/src/pages/ExamplePage.tsx)" = "frontend/src/pages/ExamplePage.test.tsx" ] || fail "page test mapping"
[ "$(page_test_for frontend/src/utils/cards/cardValue.ts)" = "frontend/src/utils/cards/cardValue.test.ts" ] || fail "utility test mapping"
[ -z "$(page_test_for frontend/src/pages/ExamplePage.test.tsx)" ] || fail "test file mapped as source"
BATCH_TESTING=1 source "$SCRIPTS/waitdone.sh"
printf '100\n' > "$tmp/start"
printf '99 101\n' > "$tmp/result"
rcg_result_is_fresh "$tmp/start" "$tmp/result" && fail "waitdone accepted a prior gate result"
printf '100 101\n' > "$tmp/result"
rcg_result_is_fresh "$tmp/start" "$tmp/result" || fail "waitdone rejected the current gate result"
echo "batch-lanes script tests passed"
