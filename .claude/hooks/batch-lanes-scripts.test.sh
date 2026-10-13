#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPTS="$ROOT/.claude/skills/batch-lanes/scripts"
fail() { echo "FAIL: $*" >&2; exit 1; }

for script in "$SCRIPTS"/*.sh; do bash -n "$script" || fail "bash -n $script"; done
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
# identgone.sh warns on rename-shaped removals and notes intentional deletions,
# while ignoring numeric fragments, comments, strings and new files.
(
  repo="$tmp/identgone"
  git init -q "$repo"
  git -C "$repo" config user.email test@example.com
  git -C "$repo" config user.name test
  cat > "$repo/source.ts" <<'EOF'
const isMobile = true;
const countMismatch = true;
const countRef = countMismatch;
const moveBar = true;
const obsoleteHelper = true;
const mobileCheck = true;
const mobileCheckUse = mobileCheck;
const behaviour = true;
const mobileState = true;
const useState = mobileState;
const jsxAfter = true;
const numberValue = 1_000;
const repeatedName = isMobile;
// commentOnlyword and stringOnlyword
const note = "quotedOnlyword";
EOF
  git -C "$repo" add source.ts && git -C "$repo" commit -qm initial
  cat > "$repo/source.ts" <<'EOF'
const isMobileViewport = true;
const repeatedName = isMobileViewport;
const countMismatchViewport = true;
const countRef = true;
const move = true;
const numberValue = 2_000;
// commentOnlyword and stringOnlyword
const note = "quotedOnlyword";
<p>Don't move</p>;
const jsxAfterRenamed = true;
const templateText = `${mobileCheckIsMobile} ${mobileCheckIsMobile}`;
const behavior = true;
const mobileStat = true;
const mobileStateful = true;
const selectionCount = mobileStat + mobileStateful;
EOF
  printf 'const brandNewName = true;\n' > "$repo/new.ts"
  output=$(bash "$SCRIPTS/identgone.sh" "$repo") || fail "identgone.sh failed"
  grep -qx 'WARN IDENT_GONE source.ts isMobile -> isMobileViewport' <<<"$output" || fail "identifier rename with matching occurrence count was not warned: $output"
  grep -qx 'NOTE IDENT_GONE source.ts countMismatch' <<<"$output" || fail "rename-shaped identifier with mismatched occurrence count was not noted: $output"
  grep -qx 'NOTE IDENT_GONE source.ts moveBar' <<<"$output" || fail "short substring candidate was not noted: $output"
  grep -qx 'NOTE IDENT_GONE source.ts obsoleteHelper' <<<"$output" || fail "identifier deletion was not noted: $output"
  grep -qx 'WARN IDENT_GONE source.ts jsxAfter -> jsxAfterRenamed' <<<"$output" || fail "JSX apostrophe hid a later identifier rename: $output"
  grep -qx 'WARN IDENT_GONE source.ts behaviour -> behavior' <<<"$output" || fail "long edit-distance rename was not warned: $output"
  grep -qx 'WARN IDENT_GONE source.ts mobileState -> mobileStat' <<<"$output" || fail "closest candidate was not selected: $output"
  ! grep -qE 'numberValue|commentOnlyword|stringOnlyword|quotedOnlyword|brandNewName' <<<"$output" || fail "identgone reported a retained, numeric, comment, string, or new-file identifier: $output"
  ig_strings="$tmp/identgone-strings"; git init -q "$ig_strings"
  printf "const label = 'singleQuotedWord here';\nconst other = \"doubleQuotedWord\";\nexport const x = label + other;\n" > "$ig_strings/a.ts"
  git -C "$ig_strings" add a.ts && git -C "$ig_strings" -c user.name=t -c user.email=t@t commit -qm a
  printf "const label = 'replacementText here';\nconst other = \"anotherText\";\nexport const x = label + other;\n" > "$ig_strings/a.ts"
  output=$(bash "$SCRIPTS/identgone.sh" "$ig_strings")
  [[ -z "$output" ]] || fail "identgone counted words inside single or double quoted strings: $output"
  output=$(ALLOW_IDENT_GONE='isMobile obsoleteHelper countMismatch moveBar behaviour jsxAfter mobileCheck mobileCheckUse mobileState useState' bash "$SCRIPTS/identgone.sh" "$repo") || fail "identgone.sh failed with allowlist"
  [ -z "$output" ] || fail "ALLOW_IDENT_GONE did not suppress warning: $output"
) || exit 1

# Deleted paths are silent; Go raw strings are ignored while Go code is scanned.
(
  repo="$tmp/identgone-go"
  git init -q "$repo"
  git -C "$repo" config user.email test@example.com
  git -C "$repo" config user.name test
  printf 'package sample\nvar oldGoName = true\nvar kept = `oldGoName`\n' > "$repo/sample.go"
  printf 'const deletedName = true\n' > "$repo/deleted.ts"
  git -C "$repo" add . && git -C "$repo" commit -qm initial
  printf 'package sample\nvar oldGoNames = true\nvar kept = `oldGoNames`\n' > "$repo/sample.go"
  rm "$repo/deleted.ts"
  output=$(bash "$SCRIPTS/identgone.sh" "$repo") || fail "identgone.sh Go/deletion check failed"
  grep -qx 'WARN IDENT_GONE sample.go oldGoName -> oldGoNames' <<<"$output" || fail "Go identifier rename was not warned: $output"
  ! grep -q deletedName <<<"$output" || fail "deleted file produced output: $output"
  mkdir "$tmp/no-python"
  ln -s "$(command -v git)" "$tmp/no-python/git"
  bash_path=$(command -v bash)
  output=$(PATH="$tmp/no-python" "$bash_path" "$SCRIPTS/identgone.sh" "$repo") || fail "missing-python path returned nonzero"
  [ "$output" = 'NOTE IDENT_GONE python3 unavailable' ] || fail "missing-python note incorrect: $output"
) || exit 1

# Regression checks for review findings in the batch lane helpers.
for script in ship.sh fixpush.sh; do
  grep -q 'PIPESTATUS' "$SCRIPTS/$script" || fail "$script does not preserve push status"
  grep -q 'push_status == 0' "$SCRIPTS/$script" || fail "$script does not stop after push failure"
  grep -q 'git diff --name-only -z HEAD' "$SCRIPTS/$script" || fail "$script does not collect paths with NUL delimiters"
done

# overlap.sh checks committed lanes only when they have an open PR, and always checks dirty lanes.
(
  repo="$tmp/overlap-repo"; root="$tmp/overlap-wts"; state="$tmp/overlap-state"
  ghbin="$tmp/overlap-bin"; mkdir -p "$repo" "$root" "$state" "$ghbin"
  cat >"$ghbin/gh" <<'GH'
#!/usr/bin/env bash
[[ "${GH_FAIL:-0}" == 1 ]] && exit 1
printf '%s\n' "${GH_OPEN_BRANCHES:-lane2}"
GH
  chmod +x "$ghbin/gh"
  git init -q "$repo"; git -C "$repo" config user.email test@example.com; git -C "$repo" config user.name test
  printf 'base\n' > "$repo/shared"; printf 'base\n' > "$repo/other"
  git -C "$repo" add .; git -C "$repo" commit -qm base; git -C "$repo" branch -M develop
  git init -q --bare "$tmp/overlap-origin.git"; git -C "$tmp/overlap-origin.git" symbolic-ref HEAD refs/heads/develop
  git -C "$repo" remote add origin "$tmp/overlap-origin.git"; git -C "$repo" push -q -u origin develop; git -C "$repo" fetch -q origin
  git -C "$repo" worktree add -qb lane1 "$root/wt-ib1" origin/develop
  git -C "$repo" worktree add -qb lane2 "$root/wt-ib2" origin/develop
  git -C "$repo" worktree add -qb lane3 "$root/wt-ib3" origin/develop
  git -C "$repo" worktree add -qb lane4 "$root/wt-ib4" origin/develop
  printf 'lane1\n' >> "$root/wt-ib1/shared"; git -C "$root/wt-ib1" commit -qam lane1
  printf 'lane2\n' >> "$root/wt-ib2/shared"; git -C "$root/wt-ib2" commit -qam lane2
  printf 'lane3\n' >> "$root/wt-ib3/shared"; git -C "$root/wt-ib3" commit -qam lane3
  out=$(PATH="$ghbin:$PATH" BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash "$SCRIPTS/overlap.sh" 1)
  grep -qx 'WARN LANE_OVERLAP shared slots=1,2' <<<"$out" || fail "overlap.sh missed committed overlap: $out"
  [[ "$out" != *'slots=1,3'* ]] || fail "overlap.sh reported merged lane without open PR: $out"
  printf 'lane1\n' > "$root/wt-ib1/other"; git -C "$root/wt-ib1" add other; git -C "$root/wt-ib1" commit -qm disjoint
  out=$(PATH="$ghbin:$PATH" BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash "$SCRIPTS/overlap.sh" 1)
  [[ "$out" != *other* ]] || fail "overlap.sh reported disjoint paths: $out"
  printf 'dirty\n' >> "$root/wt-ib1/shared"; printf 'dirty\n' >> "$root/wt-ib4/shared"
  out=$(PATH="$ghbin:$PATH" BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash "$SCRIPTS/overlap.sh" 1)
  grep -qx 'WARN LANE_OVERLAP shared slots=1,2' <<<"$out" || fail "overlap.sh missed uncommitted overlap: $out"
  grep -qx 'WARN LANE_OVERLAP shared slots=1,4' <<<"$out" || fail "overlap.sh missed dirty lane without PR: $out"
  out=$(PATH="$ghbin:$PATH" ALLOW_OVERLAP=shared BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash "$SCRIPTS/overlap.sh" 1)
  [[ -z "$out" ]] || fail "overlap.sh ignored ALLOW_OVERLAP: $out"
  out=$(PATH="$ghbin:$PATH" GH_FAIL=1 BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash "$SCRIPTS/overlap.sh" 1)
  grep -qx 'NOTE LANE_OVERLAP gh unavailable; committed lanes not checked' <<<"$out" || fail "overlap.sh missed gh failure note: $out"
  grep -qx 'WARN LANE_OVERLAP shared slots=1,4' <<<"$out" || fail "overlap.sh did not limit gh failure checks to dirty lane: $out"
  [[ "$out" != *'slots=1,2'* ]] || fail "overlap.sh checked committed lane after gh failure: $out"
  git -C "$root/wt-ib2" switch -q --detach HEAD
  out=$(PATH="$ghbin:$PATH" BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash "$SCRIPTS/overlap.sh" 1)
  [[ "$out" != *'slots=1,2'* ]] || fail "overlap.sh reported detached lane: $out"
  printf 'dirty\n' >> "$root/wt-ib2/shared"
  out=$(PATH="$ghbin:$PATH" BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash "$SCRIPTS/overlap.sh" 1)
  grep -qx 'WARN LANE_OVERLAP shared slots=1,2' <<<"$out" || fail "overlap.sh skipped a dirty detached lane: $out"
) || exit 1
grep -q 'overlap.sh" "\$1" >> "\$BATCH_STATE/rc\$1.txt"' "$SCRIPTS/rc.sh" || fail "rc.sh no longer records overlap.sh in the receipt"
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
biome="$SCRIPTS/biomeignore.sh"
move_out=$(printf '%s\n' 'diff --git a/a.ts b/a.ts' '+++ b/a.ts' '+// biome-ignore lint/suspicious/noExplicitAny' '-// biome-ignore lint/suspicious/noExplicitAny' | bash "$biome")
[ -z "$move_out" ] || fail "biome-ignore move was warned: $move_out"
grow_out=$(printf '%s\n' 'diff --git a/a.ts b/a.ts' '+++ b/a.ts' '+// biome-ignore lint/suspicious/noExplicitAny' | bash "$biome")
grep -qx 'WARN NEW_BIOME_IGNORE a.ts (+1 -0)' <<<"$grow_out" || fail "biome-ignore increase was missed: $grow_out"
allow_out=$(printf '%s\n' 'diff --git a/a.ts b/a.ts' '+++ b/a.ts' '+// biome-ignore lint/suspicious/noExplicitAny' | ALLOW_BIOME_IGNORE='a.ts' bash "$biome")
[ -z "$allow_out" ] || fail "ALLOW_BIOME_IGNORE did not suppress warning: $allow_out"
# repeattests.sh: lists added, edited and untracked Go tests per package, and warns when a
# rerun of them fails (a shuffle-dependent test would otherwise fail on develop).
(
  rt="$tmp/rt"; mkdir -p "$rt/internal/x" && cd "$rt" || exit 1
  git init -q && git config user.email t@t && git config user.name t
  printf 'module rt\n\ngo 1.21\n' > go.mod
  printf 'package x\n\nimport "testing"\n\nfunc TestKept(t *testing.T) {\n\t_ = 1\n}\n\nfunc TestEdited(t *testing.T) {\n\t_ = 1\n}\n' > internal/x/a_test.go
  git add go.mod internal/x/a_test.go && git commit -qm init
  python3 - <<'PY'
p = 'internal/x/a_test.go'
s = open(p).read()
s = s.replace('func TestEdited(t *testing.T) {\n\t_ = 1\n}', 'func TestEdited(t *testing.T) {\n\t_ = 2\n}')
s += '\nfunc TestAdded(t *testing.T) {}\n'
open(p, 'w').write(s)
PY
  printf 'package x\n\nimport (\n\t"math/rand"\n\t"testing"\n)\n\nfunc TestFlaky(t *testing.T) {\n\tif rand.Intn(2) == 0 {\n\t\tt.Fatal("unlucky deal")\n\t}\n}\n' > internal/x/b_test.go
  listed=$(bash "$SCRIPTS/repeattests.sh" --list "$rt")
  [ "$listed" = $'internal/x TestAdded\ninternal/x TestEdited\ninternal/x TestFlaky' ] || fail "repeattests.sh listed the wrong tests: $listed"
  if command -v go >/dev/null 2>&1; then
    out=$(REPEAT_TESTS_COUNT=20 bash "$SCRIPTS/repeattests.sh" "$rt")
    grep -q '^WARN DEAL_DEPENDENT internal/x ' <<<"$out" || fail "repeattests.sh did not flag a test that fails half the time: $out"
    rm internal/x/b_test.go
    out=$(REPEAT_TESTS_COUNT=20 bash "$SCRIPTS/repeattests.sh" "$rt")
    [ -z "$out" ] || fail "repeattests.sh warned on stable tests: $out"
    printf 'package x\n\nimport "testing"\n\nfunc TestBroken(t *testing.T) { undefinedName() }\n' > internal/x/c_test.go
    out=$(REPEAT_TESTS_COUNT=2 bash "$SCRIPTS/repeattests.sh" "$rt")
    grep -q '^WARN REPEAT_TESTS_ERROR internal/x ' <<<"$out" && ! grep -q DEAL_DEPENDENT <<<"$out" \
      || fail "repeattests.sh labelled a compile error as deal-dependent: $out"
    rm internal/x/c_test.go
  fi
  git rm -qf internal/x/a_test.go
  listed=$(bash "$SCRIPTS/repeattests.sh" --list "$rt")
  [ -z "$listed" ] || fail "repeattests.sh listed tests from a deleted file: $listed"
) || exit 1
# ship_pr_body: a full body (own "## " headings) must not get a second Summary,
# Closes line or footer; plain summary lines still get the standard wrapper.
(
  source "$SCRIPTS/prbody.sh"
  export CLAUDE_SESSION_URL=https://example.test/session
  count() { grep -cF -- "$1" <<<"$2" || true; }
  full=$(ship_pr_body 42 $'Closes #42\n\n## Summary\n- did it\n\n🤖 Generated with [Claude Code](https://claude.com/claude-code)\n\nhttps://example.test/session' '' '')
  [ "$(count '## Summary' "$full")" = 1 ] || fail "full body got a second ## Summary: $full"
  [ "$(count 'Closes #42' "$full")" = 1 ] || fail "full body got a second Closes line: $full"
  [ "$(count 'Generated with [Claude Code]' "$full")" = 1 ] || fail "full body got a second footer: $full"
  noref=$(ship_pr_body 42 $'## Summary\n- did it' '' '')
  [ "$(head -1 <<<"$noref")" = 'Closes #42' ] || fail "full body without a reference did not get Closes first: $noref"
  [ "$(count 'Generated with [Claude Code]' "$noref")" = 1 ] || fail "full body without a footer did not get one: $noref"
  partof=$(ship_pr_body 42 $'Part of #42\n\n## Summary\n- slice' '' '')
  [ "$(count 'Closes #42' "$partof")" = 0 ] || fail "Part of body was given a Closes line: $partof"
  refs=$(ship_pr_body 42 $'## Summary\n- slice one\n\nRefs #42 (more slices follow)' '' '')
  [ "$(count 'Closes #42' "$refs")" = 0 ] || fail "Refs body was given a Closes line: $refs"
  refsplain=$(ship_pr_body 42 '- slice one. Refs #42' '' '')
  [ "$(count 'Closes #42' "$refsplain")" = 0 ] || fail "Refs summary was given a Closes line: $refsplain"
  other=$(ship_pr_body 42 $'Closes #420\n\n## Summary' '' '')
  grep -qx 'Closes #42' <<<"$other" || fail "Closes #420 was taken as a reference to #42: $other"
  plainref=$(ship_pr_body 42 $'- did it\n\nPart of #42' '' '')
  [ "$(count 'Closes #42' "$plainref")" = 0 ] || fail "summary lines that already reference the issue got a Closes line: $plainref"
  plain=$(ship_pr_body 42 '- did it' '- [x] vitest' '')
  [ "$(count '## Summary' "$plain")" = 1 ] && [ "$(count 'Closes #42' "$plain")" = 1 ] && [ "$(count '## Test plan' "$plain")" = 1 ] || fail "summary lines lost the standard wrapper: $plain"
) || exit 1
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

# ship.sh refuses a file changed on develop after the lane branched, before staging or commit.
(
  seed="$tmp/ship-seed"; origin="$tmp/ship-origin.git"; lane="$tmp/wt-ib9"; state="$tmp/ship-state"
  mkdir -p "$seed" "$state"
  git init -q "$seed"; git -C "$seed" config user.email test@example.com; git -C "$seed" config user.name test
  printf 'initial\n' > "$seed/A"; printf 'initial\n' > "$seed/B"
  git -C "$seed" add A B; git -C "$seed" commit -qm initial; initial=$(git -C "$seed" rev-parse HEAD); git -C "$seed" branch -M develop
  git init -q --bare "$origin"; git -C "$origin" symbolic-ref HEAD refs/heads/develop
  git -C "$seed" remote add origin "$origin"; git -C "$seed" push -q origin develop
  git clone -q -b develop "$origin" "$lane"; git -C "$lane" config user.email test@example.com; git -C "$lane" config user.name test
  printf 'develop update\n' >> "$seed/A"; git -C "$seed" commit -qam 'develop updates A'; git -C "$seed" push -q origin develop
  printf 'lane update\n' >> "$lane/A"
  export BATCH_REPO="$ROOT" BATCH_WT_ROOT="$tmp" BATCH_STATE="$state" CLAIM_SKIP=1
  hash=$(bash "$SCRIPTS/treehash.sh" 9); printf '%s' "$hash" > "$state/gate9.ok"
  if bash "$SCRIPTS/ship.sh" 9 123 'test ship' 'claim-check: skip' >"$tmp/ship-overlap.out" 2>&1; then fail "ship.sh accepted overlapping stale-base file"; fi
  grep -q 'REFUSED: develop changed these files after this lane branched' "$tmp/ship-overlap.out" || fail "ship.sh did not explain stale-base overlap: $(cat "$tmp/ship-overlap.out")"
  [ -z "$(git -C "$lane" status --porcelain)" ] && fail "ship.sh staged or committed stale-base changes"
  git clone -q "$origin" "$tmp/wt-ib10"
  git -C "$tmp/wt-ib10" checkout -qb lane-disjoint "$initial"
  git -C "$tmp/wt-ib10" config user.email test@example.com; git -C "$tmp/wt-ib10" config user.name test
  printf 'lane update\n' >> "$tmp/wt-ib10/B"
  hash=$(bash "$SCRIPTS/treehash.sh" 10); printf '%s' "$hash" > "$state/gate10.ok"
  set +e
  bash "$SCRIPTS/ship.sh" 10 123 'test ship' 'claim-check: skip' >"$tmp/ship-disjoint.out" 2>&1
  ship_rc=$?
  set -e
  [ "$ship_rc" -ne 0 ] || fail "ship.sh unexpectedly completed without gh stub"
  ! grep -q 'REFUSED: develop changed these files' "$tmp/ship-disjoint.out" || fail "ship.sh rejected a disjoint file change"
  git -C "$tmp/wt-ib10" log -1 --format=%s | grep -qx 'test ship' || fail "ship.sh did not pass stale-base check and commit disjoint change"
) || exit 1

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

# combined.sh: isolate a detached scratch tree, filter open PR branches, report conflicts, and honor -k.
combined="$SCRIPTS/combined.sh"
mkdir -p "$tmp/combined-bin"
cat >"$tmp/combined-bin/gh" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "${COMBINED_BRANCHES:-}"
STUB
cat >"$tmp/combined-bin/bun" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
cat >"$tmp/combined-bin/bunx" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$COMBINED_BUNX_LOG"
exit 0
STUB
chmod +x "$tmp/combined-bin/gh" "$tmp/combined-bin/bun" "$tmp/combined-bin/bunx"

make_combined_repo() {
  local repo="$1" origin="$2" shared="$3"
  mkdir -p "$repo/frontend/node_modules" "$repo/frontend"
  printf '{"name":"fixture","version":"1.0.0"}\n' > "$repo/frontend/package.json"
  git init -q "$repo"
  git -C "$repo" config user.email test@example.com
  git -C "$repo" config user.name test
  printf 'base\n' > "$repo/$shared"
  git -C "$repo" add . && git -C "$repo" -c user.name=test -c user.email=test@example.com commit -qm base
  git -C "$repo" branch -M develop
  git init -q --bare "$origin"
  git -C "$origin" symbolic-ref HEAD refs/heads/develop
  git -C "$repo" remote add origin "$origin"
  git -C "$repo" push -q -u origin develop
}
add_combined_branch() {
  local repo="$1" spec="$2" file="$3"
  git -C "$repo" checkout -q -b "$spec" origin/develop
  printf '%s\n' "$spec" > "$repo/$file"
  git -C "$repo" add "$file" && git -C "$repo" -c user.name=test -c user.email=test@example.com commit -qm "$spec"
  git -C "$repo" push -q origin "$spec"
  git -C "$repo" checkout -q develop
}
combined_repo="$tmp/combined-repo"
make_combined_repo "$combined_repo" "$tmp/combined-origin.git" base
add_combined_branch "$combined_repo" feat/11-one feat-one
add_combined_branch "$combined_repo" chore/12-skip skip
add_combined_branch "$combined_repo" feat/13-two feat-two
git -C "$combined_repo" fetch -q origin
if PATH="$tmp/combined-bin:$PATH" BATCH_REPO="$combined_repo" BATCH_WT_ROOT="$tmp" \
  bash "$combined" frontend/e2e/guard.spec.ts >"$tmp/combined-missing.out" 2>&1; then
  fail "combined.sh accepted an unset BATCH_BRANCH_RE"
fi
grep -q BATCH_BRANCH_RE "$tmp/combined-missing.out" || fail "combined.sh did not explain missing BATCH_BRANCH_RE"
export COMBINED_BUNX_LOG="$tmp/combined-bunx.log"
PATH="$tmp/combined-bin:$PATH" BATCH_REPO="$combined_repo" BATCH_WT_ROOT="$tmp" \
  BATCH_BRANCH_RE='^feat/' COMBINED_BRANCHES=$'feat/11-one\nchore/12-skip\nfeat/13-two' \
  bash "$combined" -k -r 2 frontend/e2e/guard.spec.ts >"$tmp/combined-filter.out" || fail "combined.sh failed on clean merges"
grep -q '^MERGED: feat/11-one feat/13-two$' "$tmp/combined-filter.out" || fail "combined.sh merged an unexpected branch: $(cat "$tmp/combined-filter.out")"
[ -d "$tmp/wt-combined" ] || fail "combined.sh -k removed its worktree"
grep -q 'playwright test frontend/e2e/guard.spec.ts --repeat-each=2 --reporter=line' "$tmp/combined-bunx.log" || fail "combined.sh did not pass repeat/spec arguments"
git -C "$tmp/wt-combined" worktree remove --force "$tmp/wt-combined"

conflict_repo="$tmp/combined-conflict-repo"
make_combined_repo "$conflict_repo" "$tmp/conflict-origin.git" shared
add_combined_branch "$conflict_repo" feat/21-first shared
add_combined_branch "$conflict_repo" feat/22-conflict shared
git -C "$conflict_repo" fetch -q origin
if PATH="$tmp/combined-bin:$PATH" BATCH_REPO="$conflict_repo" BATCH_WT_ROOT="$tmp" \
  BATCH_BRANCH_RE='^feat/' COMBINED_BRANCHES=$'feat/21-first\nfeat/22-conflict' \
  bash "$combined" frontend/e2e/guard.spec.ts >"$tmp/combined-conflict.out" 2>&1; then
  fail "combined.sh returned success after a merge conflict"
fi
grep -q '^CONFLICT feat/22-conflict shared$' "$tmp/combined-conflict.out" || fail "combined.sh did not report conflict branch/files: $(cat "$tmp/combined-conflict.out")"
[ ! -e "$tmp/wt-combined" ] || fail "combined.sh did not remove scratch worktree without -k"
echo "batch-lanes script tests passed"
