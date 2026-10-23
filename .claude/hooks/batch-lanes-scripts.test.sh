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
grep -q 'adrcheck.sh" "\$1" >> "\$BATCH_STATE/rc\$1.txt"' "$SCRIPTS/rc.sh" || fail "rc.sh no longer records adrcheck.sh in the receipt"

# adrcheck.sh uses the lane's changed paths and allows orchestrator-written ADR progress lines.
(
  repo="$tmp/adrcheck"; root="$tmp/adrcheck-lanes"; state="$tmp/adrcheck-state"
  mkdir -p "$root"; git init -q --bare "$tmp/adrcheck-origin.git"; git init -q "$repo"
  git -C "$repo" config user.email test@example.com; git -C "$repo" config user.name test
  mkdir -p "$repo/docs/adr"; printf 'base\n' > "$repo/tracked"; printf 'decision\n' > "$repo/docs/adr/0001-test.md"
  git -C "$repo" add tracked docs/adr/0001-test.md; git -C "$repo" commit -qm base
  git -C "$repo" branch -M develop; git -C "$repo" remote add origin "$tmp/adrcheck-origin.git"
  git -C "$repo" push -q -u origin develop; git -C "$repo" fetch -q origin
  git -C "$repo" worktree add -qb lane "$root/wt-ib1" origin/develop
  printf 'progress\n' >> "$root/wt-ib1/docs/adr/0001-test.md"
  out=$(BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash "$SCRIPTS/adrcheck.sh" 1)
  grep -qx 'WARN ADR_EDIT docs/adr/0001-test.md' <<<"$out" || fail "adrcheck.sh missed ADR edit: $out"
  out=$(ALLOW_ADR_EDIT=1 BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash "$SCRIPTS/adrcheck.sh" 1)
  [[ -z "$out" ]] || fail "adrcheck.sh ignored ALLOW_ADR_EDIT: $out"
  printf 'decision\n' > "$root/wt-ib1/docs/adr/0001-test.md"
  out=$(BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash "$SCRIPTS/adrcheck.sh" 1)
  [[ -z "$out" ]] || fail "adrcheck.sh warned on unchanged ADR: $out"
)
! grep -R -nE 'git status --porcelain \| awk' "$SCRIPTS" || fail "status paths are parsed as whitespace fields"
if grep -R -n '/tmp/' "$SCRIPTS" >"$tmp/tmp-paths.out" 2>&1; then
  fail "batch scripts contain hardcoded /tmp paths: $(cat "$tmp/tmp-paths.out")"
fi
! grep -q 'Co-Authored-By: Claude Opus 5.5' "$SCRIPTS/ship.sh" || fail "ship.sh hardcodes co-author"
! grep -q 'gh pr view .*codecov' "$SCRIPTS/land.sh" || fail "land.sh retains result-free codecov pipeline"
grep -q 'CLAUDE_CO_AUTHOR' "$SCRIPTS/ship.sh" || fail "ship.sh does not support CLAUDE_CO_AUTHOR"
grep -q 'BATCH_DELEGATED_TO' "$SCRIPTS/ship.sh" || fail "ship.sh does not support BATCH_DELEGATED_TO"
grep -q 'Delegated-To:.*delegated_to' "$SCRIPTS/fixpush.sh" || fail "fixpush.sh does not write inferred delegated-to metadata"
grep -q 'gh pr merge .*>/dev/null' "$SCRIPTS/land.sh" || fail "land.sh does not suppress merge stdout"
grep -q 'Failed to merge PR' "$SCRIPTS/land.sh" || fail "land.sh does not report merge failures"
grep -q 'go_trumpcards 固有の規約' "$SCRIPTS/receipt.sh" || fail "receipt.sh does not explain project-specific conventions"
grep -qF '[ -f "$t.test.ts" ] || [ -f "$t.test.tsx" ]' "$SCRIPTS/receipt.sh" || fail "receipt.sh MISSING_TEST must accept a .test.tsx file for a .ts source (hooks are often tested with providers in .tsx)"
grep -qF '! "$file" =~ \.test\. && -f "$file" ]] && new_sources+=' "$SCRIPTS/receipt.sh" || fail "receipt.sh MISSING_TEST must skip a deleted source (deleting a rule mirror and its test is not a missing test)"
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
printf 'delegated\n' > "$DELE_WORKTREE/result"
STUB
chmod +x "$tmp/dele"
if PATH="$tmp:$PATH" DELE_CALLED="$tmp/called" \
  bash "$SCRIPTS/go.sh" 1 123 "$tmp/missing.md" >/dev/null 2>&1; then
  fail "go.sh accepted a missing prompt"
fi
[ ! -e "$tmp/called" ] || fail "go.sh invoked dele without a prompt"
mkdir -p "$tmp/wt-ib1"
git -C "$tmp/wt-ib1" init -q
git -C "$tmp/wt-ib1" config user.email test@example.com
git -C "$tmp/wt-ib1" config user.name test
printf 'base\n' > "$tmp/wt-ib1/base"
git -C "$tmp/wt-ib1" add base && git -C "$tmp/wt-ib1" commit -qm base
printf 'delegate this issue\n' > "$tmp/prompt.md"
PATH="$tmp:$PATH" DELE_CALLED="$tmp/called" DELE_WORKTREE="$tmp/wt-ib1" BATCH_WT_ROOT="$tmp" \
  bash "$SCRIPTS/go.sh" 1 123 "$tmp/prompt.md" >/dev/null 2>&1 || fail "go.sh rejected a non-empty prompt"
[ -e "$tmp/called" ] || fail "go.sh did not invoke dele for a valid prompt"
cat >"$tmp/dele" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
chmod +x "$tmp/dele"
mkdir -p "$tmp/empty-state"
if PATH="$tmp:$PATH" BATCH_STATE="$tmp/empty-state" BATCH_WT_ROOT="$tmp" \
  bash "$SCRIPTS/go.sh" 1 124 "$tmp/prompt.md" >"$tmp/no-change.out" 2>&1; then
  fail "go.sh accepted a successful delegation with no work"
fi
grep -q '^WARN NO_CHANGE' "$tmp/no-change.out" || fail "go.sh omitted NO_CHANGE warning: $(cat "$tmp/no-change.out")"

# The prompt template may be echoed into the delegate log; only a report whose
# first line is NOT_PLANNED can suppress NO_CHANGE.
cat >"$tmp/dele" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' 'The instructions say: write VERDICT: NOT_PLANNED only when the premise is wrong.'
printf '%s\n' 'VERDICT: IMPLEMENTED'
exit 0
STUB
chmod +x "$tmp/dele"
if PATH="$tmp:$PATH" BATCH_STATE="$tmp/empty-state" BATCH_WT_ROOT="$tmp" \
  bash "$SCRIPTS/go.sh" 1 126 "$tmp/prompt.md" >"$tmp/template-no-change.out" 2>&1; then
  fail "go.sh accepted an unchanged delegation whose prompt quoted NOT_PLANNED"
fi
grep -q '^WARN NO_CHANGE' "$tmp/template-no-change.out" || fail "go.sh suppressed NO_CHANGE because prompt text contained NOT_PLANNED: $(cat "$tmp/template-no-change.out")"

# A pre-existing modified file must still count as work when the delegate edits it again.
printf 'initial edit\n' >> "$tmp/wt-ib1/base"
cat >"$tmp/dele" <<'STUB'
#!/usr/bin/env bash
printf 'delegated edit\n' >> "$DELE_WORKTREE/base"
exit 0
STUB
chmod +x "$tmp/dele"
if PATH="$tmp:$PATH" BATCH_STATE="$tmp/empty-state" BATCH_WT_ROOT="$tmp" \
  DELE_WORKTREE="$tmp/wt-ib1" bash "$SCRIPTS/go.sh" 1 125 "$tmp/prompt.md" >"$tmp/changed-existing.out" 2>&1; then
  :
else
  fail "go.sh rejected a successful edit to an already modified file: $(cat "$tmp/changed-existing.out")"
fi
! grep -q '^WARN NO_CHANGE' "$tmp/changed-existing.out" || fail "go.sh warned NO_CHANGE after editing an already modified file"

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
git -C "$tmp/repo" worktree add -q --detach "$tmp/wt-ib21" origin/develop
if PATH="$PATH" BATCH_REPO="$tmp/repo" BATCH_WT_ROOT="$tmp" bash "$SCRIPTS/fx.sh" fx1 target >"$tmp/fx-nonnumeric.out" 2>&1; then fail "fx.sh accepted a nonnumeric slot"; else [[ $? == 2 ]] || fail "fx.sh returned the wrong status for a nonnumeric slot"; fi
grep -q 'fx slot must be numeric' "$tmp/fx-nonnumeric.out" || fail "fx.sh did not explain numeric slots"
[[ ! -e "$tmp/wt-ibfx1" ]] || fail "fx.sh created a worktree for a nonnumeric slot"
if PATH="$PATH" BATCH_REPO="$tmp/repo" BATCH_WT_ROOT="$tmp" bash "$SCRIPTS/gofx.sh" fx1 123 "$tmp/prompt" >"$tmp/gofx-nonnumeric.out" 2>&1; then fail "gofx.sh accepted a nonnumeric slot"; else [[ $? == 2 ]] || fail "gofx.sh returned the wrong status for a nonnumeric slot"; fi
grep -q 'fx slot must be numeric' "$tmp/gofx-nonnumeric.out" || fail "gofx.sh did not explain numeric slots"
PATH="$PATH" BATCH_REPO="$tmp/repo" BATCH_WT_ROOT="$tmp" \
  bash "$SCRIPTS/fx.sh" 21 target >"$tmp/fx.out" || fail "fx.sh failed to detach a clean worktree"
git -C "$tmp/other" symbolic-ref -q HEAD && fail "fx.sh left the target branch checked out elsewhere"
git -C "$tmp/other" status --porcelain | grep -q . && fail "fx.sh dirtied the detached worktree"
git -C "$tmp/wt-ib21" checkout -q --detach origin/develop
git -C "$tmp/other" checkout -q target
echo changed > "$tmp/other/file"
if PATH="$PATH" BATCH_REPO="$tmp/repo" BATCH_WT_ROOT="$tmp" \
  bash "$SCRIPTS/fx.sh" 21 target >"$tmp/fx-dirty.out" 2>&1; then fail "fx.sh accepted a dirty worktree on target"; fi
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
(
  game_repo="$tmp/receipt-games"
  mkdir -p "$game_repo/frontend/src/constants" "$game_repo/frontend/src/pages"
  printf "export const gameRoutes = [\n  { path: '/unsunkaruta', page: 'UnsunKaruta' },\n  { path: '/five-card-stud', page: 'FiveCardStud' },\n  { path: '/warlords', page: 'WarlordsWebPresenter' },\n  { path: '/war', page: 'War' },\n];\n" > "$game_repo/frontend/src/constants/gameRoutes.ts"
  touch "$game_repo/frontend/src/pages/UnsunKarutaPage.test.tsx"
  touch "$game_repo/frontend/src/pages/FiveCardStudPage.test.tsx"
  touch "$game_repo/frontend/src/pages/WarPage.test.tsx"
  touch "$game_repo/frontend/src/pages/WarlordsWebPresenterPage.test.tsx"
  selected=$(receipt_game_tests "$game_repo" feat/1-unsunkaruta-labels)
  grep -qx 'frontend/src/pages/UnsunKarutaPage.test.tsx' <<<"$selected" || fail "receipt missed route test from game branch: $selected"
  selected=$(receipt_game_tests "$game_repo" feat/1-five-card-stud-fix)
  grep -qx 'frontend/src/pages/FiveCardStudPage.test.tsx' <<<"$selected" || fail "receipt missed multiword game route test: $selected"
  RECEIPT_CHANGED=(internal/domain/warlords/WarlordsWebPresenter.go)
  selected=$(receipt_game_tests "$game_repo" feat/1-warlords-change)
  grep -qx 'frontend/src/pages/WarlordsWebPresenterPage.test.tsx' <<<"$selected" || fail "receipt did not match exact Warlords branch name: $selected"
  if [[ "$selected" == *WarPage.test.tsx* ]]; then fail "Warlords branch selected WarPage.test.tsx: $selected"; fi
  path=outer-path
  RECEIPT_CHANGED=(internal/i18n/locales/ja/Other.json)
  receipt_game_tests "$game_repo" feat/1-no-such-game >/dev/null
  [[ "$path" == outer-path ]] || fail "receipt_game_tests leaked its path variable into caller scope"
  selected=$(receipt_game_tests "$game_repo" feat/1-no-such-game)
  grep -qx 'WARN NO_TEST_MATCH' <<<"$selected" || fail "receipt did not warn on empty test selection: $selected"
  RECEIPT_CHANGED=(internal/adapter/presenter/web_presenter_i18n_guard_test.go)
  selected=$(receipt_game_tests "$game_repo" feat/1-no-such-game)
  [[ -z "$selected" ]] || fail "receipt warned on a change no page renders: $selected"
  RECEIPT_CHANGED=(internal/adapter/presenter/web_presenter_i18n_guard_test.go frontend/src/utils/x.ts)
  selected=$(receipt_game_tests "$game_repo" feat/1-no-such-game)
  grep -qx 'WARN NO_TEST_MATCH' <<<"$selected" || fail "receipt did not warn when a frontend file changed: $selected"
) || exit 1
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
# e2egate selects page runs, full sweeps, imported helper specs, and no-op diffs.
make_e2egate_case() {
  local name="$1" changed="$2" repo="$tmp/e2egate-$1" wtroot="$tmp/e2egate-wts-$1"
  mkdir -p "$repo/frontend/src/pages" "$repo/frontend/src/constants" "$repo/frontend/e2e" "$repo/internal/adapter/presenter" "$wtroot"
  git init -q "$repo"; git -C "$repo" config user.email test@example.com; git -C "$repo" config user.name test
  printf "export const gameRoutes = [{ path: '/wrong' }, { path: '/poker', page: 'Poker' }, { path: '/', page: 'BlackJack' }];\n" > "$repo/frontend/src/constants/gameRoutes.ts"
  for spec in desktop-board-visibility desktop-hand-visibility game-message-visibility mobile-viewport text-contrast route-sweep poker poker-paytable-mobile pokerx blackjack; do touch "$repo/frontend/e2e/$spec.spec.ts"; done
  mkdir -p "$repo/frontend/src/components" "$repo/frontend/src/i18n/locales/en" "$repo/frontend/src/types" "$repo/frontend/src/utils"
  printf "import { helper } from './helpers';\n" > "$repo/frontend/e2e/poker.spec.ts"
  touch "$repo/frontend/e2e/helpers.ts" "$repo/frontend/src/components/Other.tsx" "$repo/frontend/src/i18n/locales/en/common.json" "$repo/frontend/src/components/a.test.tsx"
  printf 'export interface Thing { name: string }\nexport const runtime = true;\n' > "$repo/frontend/src/types/index.ts"
  printf 'export const helper = true;\n' > "$repo/frontend/src/utils/obsolete.ts"
  touch "$repo/frontend/src/pages/PokerPage.tsx" "$repo/frontend/src/components/GameFooter.tsx" "$repo/frontend/base.ts"
  touch "$repo/frontend/src/pages/BlackJackPage.tsx" "$repo/internal/adapter/presenter/BlackjackWebPresenter.go"
  printf "import { PokerWidget } from '../components/poker/PokerWidget';\\n" > "$repo/frontend/src/pages/PokerPage.tsx"
  touch "$repo/frontend/src/pages/OtherPage.tsx"
  mkdir -p "$repo/frontend/src/components/poker"
  mkdir -p "$repo/frontend/src/components/blackjack"
  touch "$repo/frontend/src/components/blackjack/BlackjackWidget.tsx"
  touch "$repo/frontend/src/components/poker/PokerWidget.tsx" "$repo/internal/adapter/presenter/PokerWebPresenter.go" "$repo/internal/adapter/presenter/UnknownWebPresenter.go"
  git -C "$repo" add .
  tree=$(git -C "$repo" write-tree)
  commit=$(printf "base\n" | git -C "$repo" -c user.name=test -c user.email=test@example.com commit-tree "$tree")
  git -C "$repo" update-ref refs/heads/develop "$commit"; git -C "$repo" update-ref refs/remotes/origin/develop "$commit"
  git -C "$repo" worktree add -q -b lane "$wtroot/wt-ib1" "$commit"
  case "$changed" in
    shared) printf "\n// edit\n" >> "$wtroot/wt-ib1/frontend/src/components/GameFooter.tsx" ;;
    page) printf "\n// edit\n" >> "$wtroot/wt-ib1/frontend/src/pages/PokerPage.tsx" ;;
    component) printf "\n// edit\n" >> "$wtroot/wt-ib1/frontend/src/components/Other.tsx" ;;
    helper) printf "\n// edit\n" >> "$wtroot/wt-ib1/frontend/e2e/helpers.ts" ;;
    spec) printf "\n// edit\n" >> "$wtroot/wt-ib1/frontend/e2e/poker.spec.ts" ;;
    ignored) printf "\n// edit\n" >> "$wtroot/wt-ib1/frontend/src/i18n/locales/en/common.json"; printf "\n// edit\n" >> "$wtroot/wt-ib1/frontend/src/components/a.test.tsx" ;;
    missing) rm "$wtroot/wt-ib1/frontend/e2e/route-sweep.spec.ts"; printf "\n// edit\n" >> "$wtroot/wt-ib1/frontend/src/components/Other.tsx" ;;
    deleted_type_page) rm "$wtroot/wt-ib1/frontend/src/utils/obsolete.ts"; printf '\nexport type Other = string;\n' >> "$wtroot/wt-ib1/frontend/src/types/index.ts"; printf "\n// edit\n" >> "$wtroot/wt-ib1/frontend/src/pages/PokerPage.tsx" ;;
    type_runtime) sed -i 's/export const runtime = true;/export const runtime = false;/' "$wtroot/wt-ib1/frontend/src/types/index.ts" ;;
    type_large) for ((i=0; i<2200; i++)); do printf '// %s\n' 'large type-only comment padding line with enough bytes to exceed the pipe buffer' >> "$wtroot/wt-ib1/frontend/src/types/index.ts"; done; sed -i 's/export const runtime = true;/export const runtime = false;/' "$wtroot/wt-ib1/frontend/src/types/index.ts" ;;
    restored_deleted) cp "$wtroot/wt-ib1/frontend/src/utils/obsolete.ts" "$tmp/e2egate-restored-source"; rm "$wtroot/wt-ib1/frontend/src/utils/obsolete.ts"; git -C "$wtroot/wt-ib1" add frontend/src/utils/obsolete.ts; cp "$tmp/e2egate-restored-source" "$wtroot/wt-ib1/frontend/src/utils/obsolete.ts" ;;
    deleted_component) rm "$wtroot/wt-ib1/frontend/src/utils/obsolete.ts"; printf "\n// edit\n" >> "$wtroot/wt-ib1/frontend/src/components/Other.tsx" ;;
    presenter) printf '\n// edit\n' >> "$wtroot/wt-ib1/internal/adapter/presenter/PokerWebPresenter.go" ;;
    presenter_unmapped) printf '\n// edit\n' >> "$wtroot/wt-ib1/internal/adapter/presenter/UnknownWebPresenter.go" ;;
    game_component) printf '\n// edit\n' >> "$wtroot/wt-ib1/frontend/src/components/poker/PokerWidget.tsx" ;;
    game_component_other) printf "import { PokerWidget } from '../components/poker/PokerWidget';\\n" >> "$wtroot/wt-ib1/frontend/src/pages/OtherPage.tsx"; printf '\n// edit\n' >> "$wtroot/wt-ib1/frontend/src/components/poker/PokerWidget.tsx" ;;
    game_component_shared) printf "import { PokerWidget } from './poker/PokerWidget';\\n" >> "$wtroot/wt-ib1/frontend/src/components/Shared.tsx"; printf '\n// edit\n' >> "$wtroot/wt-ib1/frontend/src/components/poker/PokerWidget.tsx" ;;
    presenter_case) printf '\n// edit\n' >> "$wtroot/wt-ib1/internal/adapter/presenter/BlackjackWebPresenter.go" ;;
    component_case) printf '\n// edit\n' >> "$wtroot/wt-ib1/frontend/src/components/blackjack/BlackjackWidget.tsx" ;;
    shared_component) printf '\n// edit\n' >> "$wtroot/wt-ib1/frontend/src/components/GameFooter.tsx"; printf "import { GameFooter } from '../components/GameFooter';\\n" >> "$wtroot/wt-ib1/frontend/src/pages/OtherPage.tsx" ;;
  esac
  if [[ "$changed" == missing ]]; then
    error="$tmp/e2egate-missing.err"
    if E2E_SWEEP_REPEAT=3 BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-state-$name" bash "$SCRIPTS/e2egate.sh" --dry-run 1 >"$tmp/e2egate-missing.out" 2>"$error"; then
      fail "e2egate accepted missing route-sweep spec"
    else
      status=$?
      [[ $status == 2 ]] || fail "missing sweep spec exited $status instead of 2"
    fi
    grep -qx 'e2egate: sweep spec route-sweep.spec.ts does not exist' "$error" || fail "missing sweep diagnostic incorrect: $(cat "$error")"
    return
  fi
  output=$(E2E_SWEEP_REPEAT=3 BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-state-$name" bash "$SCRIPTS/e2egate.sh" --dry-run 1) || fail "e2egate $name dry-run failed"
  if [[ "$changed" == shared || "$changed" == component || "$changed" == type_runtime || "$changed" == type_large || "$changed" == deleted_component || "$changed" == restored_deleted || "$changed" == shared_component ]]; then
    for spec in desktop-board-visibility desktop-hand-visibility game-message-visibility mobile-viewport text-contrast route-sweep; do grep -q "e2e/$spec.spec.ts" <<< "$output" || fail "sweep omitted $spec: $output"; done
  elif [[ "$changed" == page ]]; then
    if E2E_PAGE_REPEAT=bad BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-state-$name" bash "$SCRIPTS/e2egate.sh" --dry-run 1 >"$tmp/e2egate-page-repeat.out" 2>"$tmp/e2egate-page-repeat.err"; then
      fail "e2egate accepted nonnumeric E2E_PAGE_REPEAT"
    else
      [[ $? == 2 ]] || fail "invalid E2E_PAGE_REPEAT did not exit 2"
    fi
    grep -qx 'E2E_PAGE_REPEAT must be numeric' "$tmp/e2egate-page-repeat.err" || fail "invalid E2E_PAGE_REPEAT diagnostic missing"
    grep -qF 'run 1: e2e/poker-paytable-mobile.spec.ts e2e/poker.spec.ts --repeat-each=3' <<< "$output" || fail "repeated page specs missing from separate run: $output"
    ! grep -q 'pokerx.spec.ts' <<< "$output" || fail "page route selected a spec without a hyphen boundary: $output"
    grep -qF 'run 2: e2e/desktop-board-visibility.spec.ts' <<< "$output" || fail "page sweep run omitted: $output"
    grep -qF -- '--grep /poker\b --pass-with-no-tests --repeat-each=3' <<< "$output" || fail "page sweep omitted route grep/pass-with-no-tests/repeat: $output"
    output_once=$(E2E_SWEEP_REPEAT=1 BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-state-$name" bash "$SCRIPTS/e2egate.sh" --dry-run 1) || fail "e2egate $name repeat=1 dry-run failed"
    ! grep -qF -- '--grep /poker\b --pass-with-no-tests --repeat-each' <<<"$output_once" || fail "sweep repeat=1 still added sweep repeat arg: $output_once"
    exec_scripts="$tmp/e2egate-exec-$name"
    mkdir -p "$exec_scripts"
    cp "$SCRIPTS/common.sh" "$SCRIPTS/e2egate.sh" "$exec_scripts/"
    cat > "$exec_scripts/e2e.sh" <<'STUB'
#!/usr/bin/env bash
    printf '%s\n' "$@" >> "$E2EGATE_ARGS"
    printf '%s\n' '---CALL---' >> "$E2EGATE_ARGS"
STUB
    chmod +x "$exec_scripts/e2e.sh"
    E2E_SWEEP_REPEAT=3 E2EGATE_ARGS="$tmp/e2egate-args-$name" BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-state-$name" bash "$exec_scripts/e2egate.sh" 1 || fail "page e2egate execution failed"
    expected="$tmp/e2egate-expected-$name"
    printf '%s\n' 1 e2e/poker-paytable-mobile.spec.ts e2e/poker.spec.ts --repeat-each=3 ---CALL--- 1 e2e/desktop-board-visibility.spec.ts e2e/desktop-hand-visibility.spec.ts e2e/game-message-visibility.spec.ts e2e/mobile-viewport.spec.ts e2e/route-sweep.spec.ts e2e/text-contrast.spec.ts --grep '/poker\b' --pass-with-no-tests --repeat-each=3 ---CALL--- > "$expected"
    diff -u "$expected" "$tmp/e2egate-args-$name" || fail "page execution passed wrong args"
    E2E_SWEEP_REPEAT=1 E2EGATE_ARGS="$tmp/e2egate-args-$name" BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-state-$name" bash "$exec_scripts/e2egate.sh" 1 || fail "page e2egate repeat=1 execution failed"
    grep -q -- '--repeat-each=3' "$tmp/e2egate-args-$name" || fail "page repeat was removed when sweep repeat=1"
    ! grep -q 'e2e/e2e/' "$tmp/e2egate-args-$name" || fail "page execution doubled e2e prefix"
  elif [[ "$changed" == presenter || "$changed" == game_component ]]; then
    grep -q '^run 1: e2e/poker-paytable-mobile.spec.ts e2e/poker.spec.ts --repeat-each=3$' <<< "$output" || fail "$changed did not repeat its page spec: $output"
    grep -qF -- '--grep /poker\b --pass-with-no-tests --repeat-each=3' <<< "$output" || fail "$changed did not route-filter the sweep: $output"
    ! grep -q 'pokerx.spec.ts' <<< "$output" || fail "$changed overmatched another game spec: $output"
  elif [[ "$changed" == presenter_case || "$changed" == component_case ]]; then
    grep -q '^run 1: e2e/blackjack.spec.ts --repeat-each=3$' <<< "$output" || fail "$changed did not case-insensitively select the page: $output"
    grep -qF -- '--grep /blackjack\b --pass-with-no-tests --repeat-each=3' <<< "$output" || fail "$changed did not select the blackjack route sweep: $output"
  elif [[ "$changed" == game_component_other || "$changed" == game_component_shared ]]; then
    for spec in desktop-board-visibility desktop-hand-visibility game-message-visibility mobile-viewport text-contrast route-sweep; do grep -q "e2e/$spec.spec.ts" <<< "$output" || fail "cross-game component edit omitted full sweep $spec: $output"; done
  elif [[ "$changed" == presenter_unmapped ]]; then
    [[ "$output" == 'NOTE e2egate: no page for internal/adapter/presenter/UnknownWebPresenter.go' ]] || fail "unmapped presenter note incorrect: $output"
    exec_scripts="$tmp/e2egate-exec-$name"
    mkdir -p "$exec_scripts"
    cp "$SCRIPTS/common.sh" "$SCRIPTS/e2egate.sh" "$exec_scripts/"
    printf '#!/usr/bin/env bash\nexit 0\n' > "$exec_scripts/e2e.sh"
    chmod +x "$exec_scripts/e2e.sh"
    BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-state-$name" bash "$exec_scripts/e2egate.sh" 1 >"$tmp/e2egate-real.out" 2>"$tmp/e2egate-real.err" || fail "unmapped presenter runtime failed"
    grep -qx 'NOTE e2egate: no page for internal/adapter/presenter/UnknownWebPresenter.go' "$tmp/e2egate-real.err" || fail "unmapped presenter runtime note missing: $(cat "$tmp/e2egate-real.err")"
  elif [[ "$changed" == deleted_type_page ]]; then
    grep -q '^run 1: e2e/poker-paytable-mobile.spec.ts e2e/poker.spec.ts --repeat-each=3$' <<< "$output" || fail "page game spec omitted with deleted/type-only changes: $output"
    grep -qF -- '--grep /poker\b --pass-with-no-tests --repeat-each=3' <<< "$output" || fail "page route-filtered sweep omitted for deleted/type-only changes: $output"
    [[ $(grep -c '^run ' <<< "$output") == 2 ]] || fail "deleted/type-only changes triggered an unfiltered sweep: $output"
    grep -q 'NOTE e2egate skipped full sweep for frontend/src/utils/obsolete.ts (deleted source)' <<< "$output" || fail "deleted source note omitted: $output"
    grep -q 'NOTE e2egate skipped full sweep for frontend/src/types/index.ts (type-only diff)' <<< "$output" || fail "type-only note omitted: $output"
    exec_scripts="$tmp/e2egate-exec-$name"
    mkdir -p "$exec_scripts"
    cp "$SCRIPTS/common.sh" "$SCRIPTS/e2egate.sh" "$exec_scripts/"
    printf '#!/usr/bin/env bash\nprintf "%%s\\n" "$@" > "$E2EGATE_ARGS"\n' > "$exec_scripts/e2e.sh"
    chmod +x "$exec_scripts/e2e.sh"
    E2E_SWEEP_REPEAT=1 E2EGATE_ARGS="$tmp/e2egate-args-$name" BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-state-$name" bash "$exec_scripts/e2egate.sh" 1 >"$tmp/e2egate-real.out" 2>"$tmp/e2egate-real.err" || fail "real e2egate execution failed"
    grep -q '^NOTE e2egate skipped full sweep for frontend/src/utils/obsolete.ts (deleted source)$' "$tmp/e2egate-real.err" || fail "real run omitted deleted source note: $(cat "$tmp/e2egate-real.err")"
    grep -q '^NOTE e2egate skipped full sweep for frontend/src/types/index.ts (type-only diff)$' "$tmp/e2egate-real.err" || fail "real run omitted type-only note: $(cat "$tmp/e2egate-real.err")"
  elif [[ "$changed" == helper ]]; then
    grep -q 'run 1: e2e/poker.spec.ts' <<< "$output" || fail "helper edit omitted importing spec: $output"
    [[ $(grep -c '^run ' <<< "$output") == 1 ]] || fail "helper edit selected non-importing specs: $output"
  elif [[ "$changed" == spec ]]; then
    grep -q 'run 1: e2e/poker.spec.ts$' <<< "$output" || fail "spec-only edit omitted its spec: $output"
    exec_scripts="$tmp/e2egate-exec-$name"
    mkdir -p "$exec_scripts"
    cp "$SCRIPTS/common.sh" "$SCRIPTS/e2egate.sh" "$exec_scripts/"
    cat > "$exec_scripts/e2e.sh" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$@" > "$E2EGATE_ARGS"
STUB
    chmod +x "$exec_scripts/e2e.sh"
    E2EGATE_ARGS="$tmp/e2egate-args-$name" BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-state-$name" bash "$exec_scripts/e2egate.sh" 1 || fail "spec-only e2egate execution failed"
    [[ "$(cat "$tmp/e2egate-args-$name")" == $'1\ne2e/poker.spec.ts' ]] || fail "spec-only execution passed wrong args: $(cat "$tmp/e2egate-args-$name")"
  elif [[ "$changed" == ignored ]]; then
    [[ -z "$output" ]] || fail "test/locale changes selected specs: $output"
    BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-state-$name" bash "$SCRIPTS/e2egate.sh" 1 || fail "no-spec e2egate should exit 0"
  else [[ -z "$output" ]] || fail "unchanged frontend selected specs: $output"; fi
}
make_e2egate_case shared shared
make_e2egate_case page page
make_e2egate_case spec spec
make_e2egate_case component component
make_e2egate_case helper helper
make_e2egate_case ignored ignored
make_e2egate_case empty none
make_e2egate_case missing missing
make_e2egate_case deleted-type-page deleted_type_page
make_e2egate_case types-runtime type_runtime
make_e2egate_case types-large type_large
make_e2egate_case restored-deleted restored_deleted
make_e2egate_case deleted-component deleted_component
make_e2egate_case presenter presenter
make_e2egate_case presenter-unmapped presenter_unmapped
make_e2egate_case game-component game_component
make_e2egate_case game-component-other game_component_other
make_e2egate_case game-component-shared game_component_shared
make_e2egate_case presenter-case presenter_case
make_e2egate_case component-case component_case
make_e2egate_case shared-component shared_component

# Removed route references are repeated; moves are excluded, including during full sweeps.
(
  repo="$tmp/e2egate-removed"; wtroot="$tmp/e2egate-removed-wts"
  mkdir -p "$repo/frontend/e2e" "$repo/frontend/src/components" "$wtroot"
  git -C "$repo" init -q; git -C "$repo" config user.email test@example.com; git -C "$repo" config user.name test
  for spec in desktop-board-visibility desktop-hand-visibility game-message-visibility mobile-viewport text-contrast route-sweep; do touch "$repo/frontend/e2e/$spec.spec.ts"; done
  printf "const route = '/baloot';\n" > "$repo/frontend/e2e/helpers.ts"
  printf 'base\n' > "$repo/frontend/src/components/shared.tsx"
  git -C "$repo" add .
  tree=$(git -C "$repo" write-tree)
  commit=$(printf 'base\n' | git -C "$repo" -c user.name=test -c user.email=test@example.com commit-tree "$tree")
  git -C "$repo" update-ref refs/heads/develop "$commit"; git -C "$repo" update-ref refs/remotes/origin/develop "$commit"
  git -C "$repo" worktree add -q -b lane "$wtroot/wt-ib1" "$commit"
  sed -i '/baloot/d' "$wtroot/wt-ib1/frontend/e2e/helpers.ts"
  output=$(E2E_SWEEP_REPEAT=3 BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-removed-state" bash "$SCRIPTS/e2egate.sh" --dry-run 1) || fail "removed route dry-run failed"
  grep -qF -- '--grep /baloot\b --pass-with-no-tests --repeat-each=3' <<<"$output" || fail "removed route did not get repeated: $output"
  exec_scripts="$tmp/e2egate-removed-exec"; mkdir -p "$exec_scripts"
  cp "$SCRIPTS/common.sh" "$SCRIPTS/e2egate.sh" "$exec_scripts/"
  cat > "$exec_scripts/e2e.sh" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$@" >> "$E2EGATE_ARGS"
printf '%s\n' '---CALL---' >> "$E2EGATE_ARGS"
STUB
  chmod +x "$exec_scripts/e2e.sh"
  E2E_SWEEP_REPEAT=3 E2EGATE_ARGS="$tmp/e2egate-removed-args" BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-removed-state" bash "$exec_scripts/e2egate.sh" 1 || fail "removed route execution failed"
  expected="$tmp/e2egate-removed-expected"
  printf '%s\n' 1 e2e/desktop-board-visibility.spec.ts e2e/desktop-hand-visibility.spec.ts e2e/game-message-visibility.spec.ts e2e/mobile-viewport.spec.ts e2e/route-sweep.spec.ts e2e/text-contrast.spec.ts --grep '/baloot\b' --pass-with-no-tests --repeat-each=3 ---CALL--- > "$expected"
  diff -u "$expected" "$tmp/e2egate-removed-args" || fail "removed route execution passed wrong args"
  printf 'common\n' >> "$wtroot/wt-ib1/frontend/src/components/shared.tsx"
  full=$(E2E_SWEEP_REPEAT=3 BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-removed-state" bash "$SCRIPTS/e2egate.sh" --dry-run 1) || fail "full sweep with removed route dry-run failed"
  [[ $(grep -c '^run ' <<<"$full") == 2 ]] || fail "full sweep did not add separate repeated run: $full"
  grep -q '^run 1: .*desktop-board-visibility' <<<"$full" || fail "full sweep run missing: $full"
  grep -qF -- '--grep /baloot\b --pass-with-no-tests --repeat-each=3' <<<"$full" || fail "full sweep omitted removed-route run: $full"
  cp "$wtroot/wt-ib1/frontend/e2e/helpers.ts" "$tmp/e2egate-helper-backup"
  printf "const route = '/baloot';\n" >> "$wtroot/wt-ib1/frontend/e2e/helpers.ts"
  sed -i '1i test("other", async () => {});' "$wtroot/wt-ib1/frontend/e2e/helpers.ts"
  cp "$repo/frontend/src/components/shared.tsx" "$wtroot/wt-ib1/frontend/src/components/shared.tsx"
  moved=$(E2E_SWEEP_REPEAT=3 BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-removed-state" bash "$SCRIPTS/e2egate.sh" --dry-run 1) || fail "moved route dry-run failed"
  ! grep -q -- '--grep /baloot\b' <<<"$moved" || fail "route move was incorrectly repeated: $moved"
  sed -i '/baloot/d' "$wtroot/wt-ib1/frontend/e2e/helpers.ts"
  printf "const route = '/baloot';\n" > "$wtroot/wt-ib1/frontend/e2e/other.ts"
  moved_file=$(E2E_SWEEP_REPEAT=3 BATCH_REPO="$repo" BATCH_WT_ROOT="$wtroot" BATCH_STATE="$tmp/e2egate-removed-state" bash "$SCRIPTS/e2egate.sh" --dry-run 1) || fail "cross-file route move dry-run failed"
  grep -qF -- '--grep /baloot\b --pass-with-no-tests --repeat-each=3' <<<"$moved_file" || fail "cross-file route move was not repeated: $moved_file"
) || exit 1

# fieldcheck matches every changed page to its own controller, independent of branch name.
(
  repo="$tmp/fieldcheck-repo"; mkdir -p "$repo/frontend/src/pages" "$repo/internal/adapter/controller"
  git -C "$repo" init -q; git -C "$repo" config user.email test@example.com; git -C "$repo" config user.name test
  cat > "$repo/internal/adapter/controller/FooWebController.go" <<'GO'
package controller
type Foo struct { FooField string `json:"fooField"` }
GO
  cat > "$repo/internal/adapter/controller/BarWebController.go" <<'GO'
package controller
type Bar struct { BarField string `json:"barField"` }
GO
  printf 'export const FooPage = () => null;\n' > "$repo/frontend/src/pages/FooPage.tsx"
  printf 'export const BarPage = () => null;\n' > "$repo/frontend/src/pages/BarPage.tsx"
  git -C "$repo" add .
  tree=$(git -C "$repo" write-tree)
  commit=$(printf 'base\n' | git -C "$repo" -c user.name=test -c user.email=test@example.com commit-tree "$tree")
  git -C "$repo" update-ref refs/heads/unrelated-branch "$commit"; git -C "$repo" symbolic-ref HEAD refs/heads/unrelated-branch
  printf 'const x = state.fooField;\n' >> "$repo/frontend/src/pages/FooPage.tsx"
  printf 'const x = state.barField;\n' >> "$repo/frontend/src/pages/BarPage.tsx"
  output=$(bash "$SCRIPTS/fieldcheck.sh" "$repo") || fail "fieldcheck valid two-page run failed"
  [[ -z "$output" ]] || fail "fieldcheck warned on fields in their own controllers: $output"
  printf 'const x = state.fooField;\n' >> "$repo/frontend/src/pages/BarPage.tsx"
  output=$(bash "$SCRIPTS/fieldcheck.sh" "$repo") || fail "fieldcheck cross-controller run failed"
  [[ $(grep -c '^FIELDCHECK WARN:' <<<"$output") == 1 ]] || fail "fieldcheck accepted a field from another page controller: $output"
  grep -q 'state.fooField' <<<"$output" || fail "fieldcheck warning omitted cross-controller field: $output"
  printf 'const x = state.missingField;\n' >> "$repo/frontend/src/pages/BarPage.tsx"
  output=$(bash "$SCRIPTS/fieldcheck.sh" "$repo") || fail "fieldcheck missing-field run failed"
  [[ $(grep -c '^FIELDCHECK WARN:' <<<"$output") == 2 ]] || fail "fieldcheck did not emit both expected warnings: $output"
  grep -q 'state.missingField' <<<"$output" || fail "fieldcheck warning omitted missing field: $output"
) || exit 1

# stopsrv.sh targets only the requested lane port and handles empty/invalid slots.
stop_server_pid=
trap '[[ -z "$stop_server_pid" ]] || kill "$stop_server_pid" 2>/dev/null || true; rm -rf "$tmp"' EXIT
python3 -u -m http.server 18097 --bind 127.0.0.1 >"$tmp/http-server.log" 2>&1 & stop_server_pid=$!
stop_test_fallback=0
for _ in {1..20}; do
  grep -q 'Serving HTTP on' "$tmp/http-server.log" && break
  kill -0 "$stop_server_pid" 2>/dev/null || break
  sleep 0.1
done
if ! grep -q 'Serving HTTP on' "$tmp/http-server.log"; then
  # The sandbox can prohibit socket creation; still exercise PID targeting below.
  kill "$stop_server_pid" 2>/dev/null || true
  wait "$stop_server_pid" 2>/dev/null || true
  sleep 60 & stop_server_pid=$!
  stop_test_fallback=1
fi
stop_path="$PATH"
ss_probe=$(ss -ltnpH 'sport = :18097' 2>&1 || true)
if (( stop_test_fallback )) || ! grep -q 'LISTEN' <<<"$ss_probe"; then
  mkdir -p "$tmp/stopsrv-bin"
  cat > "$tmp/stopsrv-bin/ss" <<'SS'
#!/usr/bin/env bash
printf 'LISTEN 0 5 127.0.0.1:18097 0.0.0.0:* users:(("python3",pid=%s,fd=3))\n' "$STOPSERV_TEST_PID"
SS
  chmod +x "$tmp/stopsrv-bin/ss"
  stop_path="$tmp/stopsrv-bin:$PATH"
fi
stop_out=$(PATH="$stop_path" STOPSERV_TEST_PID="$stop_server_pid" BATCH_REPO="$ROOT" BATCH_STATE="$tmp/stopsrv-state" bash "$SCRIPTS/stopsrv.sh" 97) || fail "stopsrv.sh failed to stop a lane server"
grep -q 'stopped port=18097 pid=' <<<"$stop_out" || fail "stopsrv.sh output omitted the stopped process: $stop_out"
for _ in {1..50}; do kill -0 "$stop_server_pid" 2>/dev/null || break; sleep 0.1; done
! kill -0 "$stop_server_pid" 2>/dev/null || fail "stopsrv.sh left the server process alive"
wait "$stop_server_pid" 2>/dev/null || true
stop_server_pid=
[[ $(BATCH_REPO="$ROOT" BATCH_STATE="$tmp/stopsrv-state" bash "$SCRIPTS/stopsrv.sh" 98) == none ]] || fail "stopsrv.sh did not report none for an idle slot"
if BATCH_REPO="$ROOT" BATCH_STATE="$tmp/stopsrv-state" bash "$SCRIPTS/stopsrv.sh" nope >"$tmp/stopsrv-invalid.out" 2>&1; then fail "stopsrv.sh accepted a nonnumeric slot"; else [[ $? == 2 ]] || fail "stopsrv.sh returned wrong status for invalid slot"; fi

# e2e.sh splits --- into separate bunx calls, builds once, and reports failure after all calls.
e2e_repo="$tmp/e2e-repo"; e2e_wts="$tmp/e2e-wts"; e2e_bin="$tmp/e2e-bin"
mkdir -p "$e2e_repo/frontend" "$e2e_wts/wt-ib1/frontend" "$e2e_bin"
git -C "$e2e_repo" init -q
cat > "$e2e_bin/bunx" <<'SH'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$E2E_CALL_LOG"
[[ "$*" != *second* ]]
SH
chmod +x "$e2e_bin/bunx"
export E2E_CALL_LOG="$tmp/e2e-calls.log"
if PATH="$e2e_bin:$PATH" BATCH_REPO="$e2e_repo" BATCH_WT_ROOT="$e2e_wts" BATCH_STATE="$tmp/e2e-state" E2E_SKIP_BUILD=1 bash "$SCRIPTS/e2e.sh" 1 first --- second third; then fail "e2e.sh hid a failed run"; fi
[[ $(wc -l < "$E2E_CALL_LOG") == 2 ]] || fail "e2e.sh did not make two bunx calls: $(cat "$E2E_CALL_LOG")"
grep -q 'first' "$E2E_CALL_LOG" && grep -q 'second third' "$E2E_CALL_LOG" || fail "e2e.sh split run args incorrectly: $(cat "$E2E_CALL_LOG")"
# prbody recognizes every GitHub closing keyword and issue references precisely.
(
  source "$SCRIPTS/prbody.sh"
  prbody_has_ref 5 'to fix #5' || fail 'prbody_has_ref missed fix #5'
  prbody_has_ref 5 'fixed #5' || fail 'prbody_has_ref missed fixed #5'
  prbody_has_ref 5 'resolve #5' || fail 'prbody_has_ref missed resolve #5'
  prbody_has_ref 5 'Part of #5' || fail 'prbody_has_ref missed Part of #5'
  prbody_has_ref 5 'Refs #5' || fail 'prbody_has_ref missed Refs #5'
  if prbody_has_ref 5 'Closes #50'; then fail 'prbody_has_ref accepted a longer issue number'; fi
  if prbody_has_ref 5 'disclosed #5'; then fail 'prbody_has_ref accepted disclosed #5'; fi
  if prbody_has_ref 5 'prefix #5'; then fail 'prbody_has_ref accepted prefix #5'; fi
) || exit 1

# fixpush validates the PR before committing and records a pushed PR exactly once.
(
  seed="$tmp/fix-seed"; origin="$tmp/fix-origin.git"; wts="$tmp/fix-wts"; lane="$wts/wt-ib31"; state="$tmp/fix-state"; bin="$tmp/fix-bin"
  mkdir -p "$seed" "$wts" "$state" "$bin"
  git init -q "$seed"; git -C "$seed" config user.email test@example.com; git -C "$seed" config user.name test
  echo base > "$seed/file"; git -C "$seed" add file; git -C "$seed" commit -qm base; git -C "$seed" branch -M develop
  git init -q --bare "$origin"; git -C "$origin" symbolic-ref HEAD refs/heads/develop
  git -C "$seed" remote add origin "$origin"; git -C "$seed" push -qu origin develop
  git clone -q "$origin" "$lane"; git -C "$lane" config user.email test@example.com; git -C "$lane" config user.name test
  git -C "$lane" checkout -qb feat/31-fix; git -C "$lane" push -qu origin feat/31-fix
  cat > "$bin/gh" <<'GH'
#!/usr/bin/env bash
case "$*" in
  'pr view 31 --json state,headRefName -q [.state, .headRefName][]') printf '%s\n%s\n' "${FIX_STATE:-OPEN}" "${FIX_HEAD:-feat/31-fix}" ;;
  'pr comment 31 --body '*) echo comment >> "$FIX_COMMENTS" ;;
  *) echo "unexpected gh invocation: $*" >&2; exit 2 ;;
esac
GH
  chmod +x "$bin/gh"; export FIX_COMMENTS="$tmp/fix-comments"; : > "$state/reviewed.txt"
  run_fix() { PATH="$bin:$PATH" BATCH_REPO="$ROOT" BATCH_WT_ROOT="$wts" BATCH_STATE="$state" CLAIM_SKIP=1 bash "$SCRIPTS/fixpush.sh" 31 31 'fix change' 'reviewed'; }
  printf changed >> "$lane/file"
  hash=$(BATCH_REPO="$ROOT" BATCH_WT_ROOT="$wts" bash "$SCRIPTS/treehash.sh" 31); printf '%s' "$hash" > "$state/gate31.ok"
  before=$(git -C "$lane" rev-parse HEAD)
  if FIX_STATE=MERGED run_fix >"$tmp/fix-merged.out" 2>&1; then fail 'fixpush accepted a merged PR'; fi
  grep -q 'follow-up PR' "$tmp/fix-merged.out" || fail "fixpush merged refusal missing: $(cat "$tmp/fix-merged.out")"
  [ "$(git -C "$lane" rev-parse HEAD)" = "$before" ] || fail 'fixpush committed a merged PR'
  if FIX_HEAD=other-branch run_fix >"$tmp/fix-branch.out" 2>&1; then fail 'fixpush accepted a different PR head branch'; fi
  grep -q 'does not match PR' "$tmp/fix-branch.out" || fail "fixpush branch refusal missing: $(cat "$tmp/fix-branch.out")"
  [ "$(git -C "$lane" rev-parse HEAD)" = "$before" ] || fail 'fixpush committed on a mismatched branch'
  run_fix >"$tmp/fix-success.out" 2>&1 || fail "fixpush failed: $(cat "$tmp/fix-success.out")"
  [ "$(cat "$state/reviewed.txt")" = 31 ] || fail 'fixpush did not record PR 31 once'
  echo second-change >> "$lane/file"
  hash=$(BATCH_REPO="$ROOT" BATCH_WT_ROOT="$wts" bash "$SCRIPTS/treehash.sh" 31); printf '%s' "$hash" > "$state/gate31.ok"
  run_fix >"$tmp/fix-success2.out" 2>&1 || fail "fixpush second run failed: $(cat "$tmp/fix-success2.out")"
  [ "$(wc -l < "$state/reviewed.txt")" -eq 1 ] || fail 'fixpush duplicated PR in reviewed.txt'
  echo third-change >> "$lane/file"
  : > "$state/reviewed.txt"
  echo pushed >> "$lane/file"
  hash=$(BATCH_REPO="$ROOT" BATCH_WT_ROOT="$wts" bash "$SCRIPTS/treehash.sh" 31); printf '%s' "$hash" > "$state/gate31.ok"
  real_git=$(command -v git); mkdir "$tmp/fix-failbin"
  cat > "$tmp/fix-failbin/git" <<'GIT'
#!/usr/bin/env bash
if [[ "$1" == push ]]; then exit 1; fi
exec "$REAL_GIT" "$@"
GIT
  chmod +x "$tmp/fix-failbin/git"
  before=$(git -C "$lane" rev-parse HEAD)
  if PATH="$tmp/fix-failbin:$bin:$PATH" REAL_GIT="$real_git" BATCH_REPO="$ROOT" BATCH_WT_ROOT="$wts" BATCH_STATE="$state" CLAIM_SKIP=1 bash "$SCRIPTS/fixpush.sh" 31 31 'failed push' reviewed >"$tmp/fix-push-fail.out" 2>&1; then fail 'fixpush accepted a failed push'; fi
  [ "$(wc -l < "$state/reviewed.txt")" -eq 0 ] || fail 'fixpush recorded reviewed PR after push failure'
) || exit 1

# ship issue/body and branch guards preserve the specified commit message behavior.
(
  mkdir -p "$tmp/ship-bin"
  cat > "$tmp/ship-bin/gh" <<'GH'
#!/usr/bin/env bash
echo "$*" >> "$SHIP_GH_LOG"
GH
  chmod +x "$tmp/ship-bin/gh"; export SHIP_GH_LOG="$tmp/ship-gh.log"
  ship_case() {
    local name="$1" branch="$2" body="$3" allow_unmentioned="${4:-}" allow_branch="${5:-}" origin="$tmp/ship-$1-origin.git" lane="$tmp/wt-ib$1" state="$tmp/ship-$1-state"
    mkdir -p "$state"; git init -q --bare "$origin"; git --git-dir="$origin" symbolic-ref HEAD refs/heads/develop; git init -q "$tmp/ship-$1-seed"
    git -C "$tmp/ship-$1-seed" config user.email test@example.com; git -C "$tmp/ship-$1-seed" config user.name test
    echo base > "$tmp/ship-$1-seed/file"; git -C "$tmp/ship-$1-seed" add file; git -C "$tmp/ship-$1-seed" commit -qm base; git -C "$tmp/ship-$1-seed" branch -M develop
    git -C "$tmp/ship-$1-seed" remote add origin "$origin"; git -C "$tmp/ship-$1-seed" push -qu origin develop
    git clone -q "$origin" "$lane"; git -C "$lane" checkout -qb "$branch"; git -C "$lane" config user.email test@example.com; git -C "$lane" config user.name test
    echo change >> "$lane/file"
    local h; h=$(BATCH_REPO="$ROOT" BATCH_WT_ROOT="$tmp" bash "$SCRIPTS/treehash.sh" "$1"); printf '%s' "$h" > "$state/gate$1.ok"
    local before; before=$(git -C "$lane" rev-parse HEAD)
    local out="$tmp/ship-$1.out" rc=0
    env PATH="$tmp/ship-bin:$PATH" BATCH_REPO="$ROOT" BATCH_WT_ROOT="$tmp" BATCH_STATE="$state" CLAIM_SKIP=1 ALLOW_STALE_BASE=1 ALLOW_UNMENTIONED_ISSUE="$allow_unmentioned" ALLOW_BRANCH_ISSUE_MISMATCH="$allow_branch" bash "$SCRIPTS/ship.sh" "$1" 11851 ship "$body" >"$out" 2>&1 || rc=$?
    echo "$rc" > "$tmp/ship-$1.rc"
    if [ "$rc" -ne 0 ]; then [ "$(git -C "$lane" rev-parse HEAD)" = "$before" ] || fail "ship $name committed despite refusal"; fi
  }
  ship_case 41 feat/11851-test 'Refs #11837'
  grep -q 'mentions another issue' "$tmp/ship-41.out" || fail 'ship did not refuse another issue reference'
  ship_case 42 feat/11851-test 'Refs #11837' 1
  [ "$(cat "$tmp/ship-42.rc")" = 0 ] || fail 'ship rejected ALLOW_UNMENTIONED_ISSUE'
  ship_case 43 feat/11851-test 'Refs #11851'
  [ "$(cat "$tmp/ship-43.rc")" = 0 ] || fail 'ship rejected matching issue reference'
  ! git -C "$tmp/wt-ib43" log -1 --format=%B | grep -q 'Closes #11851' || fail 'ship added Closes despite matching Refs'
  ship_case 44 feat/11851-test 'no issue reference'
  git -C "$tmp/wt-ib44" log -1 --format=%B | grep -q 'Closes #11851' || fail 'ship omitted default Closes line'
  ship_case 45 feat/11851-test 'see #11851; to fix #11729'
  grep -q 'WARN: body closes #11729' "$tmp/ship-45.out" || fail 'ship omitted warning for a different closed issue'
  ship_case 46 feat/11824-x 'summary'
  grep -q 'does not match' "$tmp/ship-46.out" || fail 'ship did not refuse branch issue mismatch'
  ship_case 47 feat/11824-x 'summary' '' 1
  [ "$(cat "$tmp/ship-47.rc")" = 0 ] || fail 'ship rejected ALLOW_BRANCH_ISSUE_MISMATCH'
) || exit 1

# Regressions from the batch-lanes handoff.
(
  idrepo="$tmp/ident-retained"; mkdir -p "$idrepo"; git init -q "$idrepo"
  git -C "$idrepo" config user.email test@example.com; git -C "$idrepo" config user.name test
  printf 'const cardToOutputWithFace = true;\n' > "$idrepo/a.ts"
  git -C "$idrepo" add . && git -C "$idrepo" commit -qm base
  printf 'const cardsToOutputWithFace = true;\nconst retained = cardToOutputWithFace;\n' > "$idrepo/a.ts"
  out=$(bash "$SCRIPTS/identgone.sh" "$idrepo")
  [[ "$out" != *'WARN IDENT_GONE a.ts cardToOutputWithFace'* ]] || fail "retained old identifier was warned: $out"
) || exit 1
(
  mkdir -p "$tmp/e2e-wts/wt-ib88/frontend" "$tmp/e2e-bin" "$tmp/e2e-state"
  cat > "$tmp/e2e-bin/bunx" <<'BIN'
#!/usr/bin/env bash
printf '<%s>\n' "$@" > "$E2E_ARGS_FILE"
BIN
  chmod +x "$tmp/e2e-bin/bunx"
  E2E_SKIP_BUILD=1 E2E_ARGS_FILE="$tmp/e2e-args" BATCH_REPO="$ROOT" BATCH_WT_ROOT="$tmp/e2e-wts" BATCH_STATE="$tmp/e2e-state" PATH="$tmp/e2e-bin:$PATH" bash "$SCRIPTS/e2e.sh" 88 '--grep=two words'
  grep -qx '<--grep=two words>' "$tmp/e2e-args" || fail "e2e.sh split a spaced Playwright argument: $(cat "$tmp/e2e-args")"
) || exit 1
(
  [[ -x "$SCRIPTS/mutate.sh" ]] || fail 'mutate.sh is missing'
) || exit 1
(
  repo="$tmp/delegation-repo"; state="$tmp/delegation-state"; root="$tmp/delegation-wts"
  mkdir -p "$repo" "$state" "$root/wt-ib9"; git init -q "$repo"
  git -C "$repo" config user.email test@example.com; git -C "$repo" config user.name test
  printf 'base\n' > "$repo/file"; git -C "$repo" add file; git -C "$repo" commit -qm base
  git -C "$repo" checkout -qb feat/current
  git -C "$repo" switch -q -c feat/main
  git -C "$root/wt-ib9" init -q
  git -C "$root/wt-ib9" config user.email test@example.com; git -C "$root/wt-ib9" config user.name test
  printf 'lane\n' > "$root/wt-ib9/file"; git -C "$root/wt-ib9" add file; git -C "$root/wt-ib9" commit -qm lane
  git -C "$root/wt-ib9" switch -q -c feat/current
  printf 'feat/current\n' > "$state/slot9.branch"
  touch -t 202001010000 "$state/slot9.branch"
  printf 'dele: old-backend ok\n' > "$state/go-9-old.log"; touch -t 201901010000 "$state/go-9-old.log"
  printf 'dele: new-backend ok\n' > "$state/go-9-new.log"
  out=$(BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash -c 'source "$1"; delegated_to 9' _ "$SCRIPTS/common.sh")
  [ "$out" = new-backend ] || fail "delegated_to included old log or missed new log: $out"
  printf 'feat/other\n' > "$state/slot9.branch"
  out=$(BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash -c 'source "$1"; delegated_to 9' _ "$SCRIPTS/common.sh")
  [ "$out" = unknown ] || fail "delegated_to did not report unknown for mismatched branch: $out"
  printf 'feat/current\n' > "$state/slot9.branch"
  touch -t 202001010000 "$state/slot9.branch"
  out=$(BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash -c 'source "$1"; delegated_to 9' _ "$SCRIPTS/common.sh")
  [ "$out" = new-backend ] || fail "delegated_to did not use lane branch when repo branch differs: $out"
) || exit 1
(
  repo="$tmp/mutate-repo"; state="$tmp/mutate-state"; root="$tmp/mutate-wts"
  mkdir -p "$repo" "$state" "$root/wt-ib9"; git init -q "$root/wt-ib9"
  git -C "$root/wt-ib9" config user.email test@example.com; git -C "$root/wt-ib9" config user.name test
  printf 'original\n' > "$root/wt-ib9/target"
  git -C "$root/wt-ib9" add target; git -C "$root/wt-ib9" commit -qm original
  printf 'mutated\n' > "$root/wt-ib9/target"
  git -C "$root/wt-ib9" add target; git -C "$root/wt-ib9" commit -qm mutation
  mutation_ref=$(git -C "$root/wt-ib9" rev-parse HEAD)
  git -C "$root/wt-ib9" switch -q --detach HEAD~1
  cat > "$tmp/fake-e2e.sh" <<'E2E'
#!/usr/bin/env bash
[[ "${E2E_ALWAYS_FAIL:-0}" == 1 ]] && exit 1
if [[ "$(cat "$BATCH_WT_ROOT/wt-ib$1/target")" == mutated && "${E2E_MUTATION_FAIL:-0}" == 1 ]]; then exit 1; fi
exit 0
E2E
  chmod +x "$tmp/fake-e2e.sh"
  run_mutate() {
    BATCH_REPO="$root/wt-ib9" BATCH_WT_ROOT="$root" BATCH_STATE="$state" BATCH_E2E_SCRIPT="$tmp/fake-e2e.sh" \
      E2E_ALWAYS_FAIL="${E2E_ALWAYS_FAIL:-0}" E2E_MUTATION_FAIL="${E2E_MUTATION_FAIL:-0}" \
      bash "$SCRIPTS/mutate.sh" 9 "$mutation_ref:target"
  }
  if E2E_ALWAYS_FAIL=1 run_mutate >"$tmp/mutate-baseline.out" 2>&1; then fail 'mutate.sh accepted a failing baseline'; fi
  grep -qx 'ERROR baseline failed' "$tmp/mutate-baseline.out" || fail "baseline failure was not reported: $(cat "$tmp/mutate-baseline.out")"
  [ "$(cat "$root/wt-ib9/target")" = original ] || fail 'mutate.sh changed target after baseline failure'
  out=$(E2E_MUTATION_FAIL=1 run_mutate) || fail "mutate.sh failed mutation-detection case: $out"
  grep -q "^backup: $state/mutate\." <<<"$out" || fail "mutate.sh did not print backup path: $out"
  grep -qx DETECTED <<<"$out" || fail "mutate.sh did not report DETECTED: $out"
  [ "$(cat "$root/wt-ib9/target")" = original ] || fail 'mutate.sh did not restore target after detection'
  out=$(run_mutate) || fail "mutate.sh failed non-detection case: $out"
  grep -qx 'NOT DETECTED' <<<"$out" || fail "mutate.sh did not report NOT DETECTED: $out"
  [ "$(cat "$root/wt-ib9/target")" = original ] || fail 'mutate.sh did not restore target after non-detection'
) || exit 1
grep -q 'rcg\$slot.pid' "$SCRIPTS/rcg.sh" || fail 'rcg.sh does not track per-slot gate pid'
(
  state="$tmp/rcg-reuse-state"; mkdir -p "$state"
  sleep 60 & unrelated_pid=$!
  printf '%s\n' "$unrelated_pid" > "$state/rcg98.pid"
  BATCH_REPO="$ROOT" BATCH_STATE="$state" BATCH_WT_ROOT="$tmp" E2E_GATE=0 bash "$SCRIPTS/rcg.sh" 98 >/dev/null 2>&1 || true
  if ! kill -0 "$unrelated_pid" 2>/dev/null; then fail 'rcg.sh killed unrelated process whose PID was reused'; fi
  kill "$unrelated_pid" 2>/dev/null || true; wait "$unrelated_pid" 2>/dev/null || true
) || exit 1
grep -q 'delegated_to "\$s"' "$SCRIPTS/ship.sh" || fail 'ship.sh does not infer Delegated-To from lane logs'
! grep -q '\[\[ -n "\$delegation" \]\]' "$SCRIPTS/ship.sh" || fail 'ship.sh retains dead empty-delegation check'
grep -q 'delegated_to "\$1"' "$SCRIPTS/fixpush.sh" || fail 'fixpush.sh does not infer Delegated-To from lane logs'


# Shared file collection, stale-base overlap, and lane gating behavior.
(
  repo="$tmp/gating-repo"; root="$tmp/gating-wts"; state="$tmp/gating-state"; bin="$tmp/gating-bin"
  mkdir -p "$root/wt-ib7" "$state" "$bin"
  git init -q "$repo"; git -C "$repo" config user.email t@t; git -C "$repo" config user.name t
  printf 'base\n' > "$repo/shared"; printf 'base\n' > "$repo/elsewhere"
  git -C "$repo" add .; git -C "$repo" commit -qm base; git -C "$repo" branch -M develop
  git init -q --bare "$tmp/gating-origin.git"; git -C "$tmp/gating-origin.git" symbolic-ref HEAD refs/heads/develop
  git -C "$repo" remote add origin "$tmp/gating-origin.git"; git -C "$repo" push -q -u origin develop; git -C "$repo" fetch -q origin
  git -C "$repo" worktree add -qb lane "$root/wt-ib7" origin/develop
  printf 'dirty\n' > "$root/wt-ib7/shared"
  cat > "$bin/gh" <<'GH'
#!/usr/bin/env bash
if [[ "$*" == *'--json files'* ]]; then printf '%s\n' "${PR_FILES:-shared}"; elif [[ "$*" == *headRefName* ]]; then echo lane; else exit 0; fi
GH
  chmod +x "$bin/gh"
  cat > "$root/wt-ib7/rcg.sh" <<'RG'
#!/usr/bin/env bash
sleep 60
RG
  chmod +x "$root/wt-ib7/rcg.sh"
  "$root/wt-ib7/rcg.sh" 7 & gatepid=$!
  trap 'kill "$gatepid" 2>/dev/null || true; wait "$gatepid" 2>/dev/null || true' EXIT
  printf '%s\n' "$gatepid" > "$state/rcg7.pid"; printf '123\n' > "$state/reviewed.txt"
  # Use an isolated merge gate; overlap returns before any merge gate work.
  out=''; rc=0
  PATH="$bin:$PATH" BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" BATCH_SKIP_CODECOV=1 bash "$SCRIPTS/land.sh" 123 >"$tmp/land.out" 2>&1 || rc=$?
  out=$(cat "$tmp/land.out")
  [[ "$rc" = 4 ]] || fail "land.sh did not return 4 for live overlap (rc=$rc): $out"
  grep -qx 'LANE-GATING slot=7 shared' <<<"$out" || fail "land.sh omitted LANE-GATING line: $out"
  rc=0
  out=$(PATH="$bin:$PATH" PR_FILES=elsewhere BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" BATCH_SKIP_CODECOV=1 bash "$SCRIPTS/land.sh" 123 2>&1) || rc=$?
  [[ "$rc" != 4 ]] || fail "land.sh returned 4 for a disjoint file: $out"
  [[ "$out" != *LANE-GATING* ]] || fail "land.sh reported a disjoint file: $out"
  out=$(PATH="$bin:$PATH" BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" BATCH_SKIP_CODECOV=1 ALLOW_LANE_GATING=1 bash "$SCRIPTS/land.sh" 123 2>&1) || true
  [[ "$out" != *LANE-GATING* ]] || fail "ALLOW_LANE_GATING did not bypass check: $out"
  kill "$gatepid" 2>/dev/null || true; wait "$gatepid" 2>/dev/null || true
  printf '%s\n' 99999999 > "$state/rcg7.pid"
  out=$(PATH="$bin:$PATH" BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" BATCH_SKIP_CODECOV=1 bash "$SCRIPTS/land.sh" 123 2>&1) || true
  [[ "$out" != *LANE-GATING* ]] || fail "land.sh treated dead gate as live: $out"
  # Develop changed shared after branch point; only lane overlap is returned.
  printf 'develop\n' >> "$repo/shared"; printf 'develop\n' >> "$repo/elsewhere"
  git -C "$repo" add .; git -C "$repo" commit -qm develop-change; git -C "$repo" push -q origin develop; git -C "$root/wt-ib7" fetch -q origin
  printf 'lane\n' > "$root/wt-ib7/shared"; printf 'lane\n' > "$root/wt-ib7/local-only"
  stale=$(BATCH_REPO="$repo" BATCH_WT_ROOT="$root" BATCH_STATE="$state" bash -c 'source "$1"; stale_base_files "$2"' _ "$SCRIPTS/common.sh" "$root/wt-ib7")
  grep -qx shared <<<"$stale" || fail "stale-base overlap missing: $stale"
  [[ "$stale" != *elsewhere* && "$stale" != *local-only* ]] || fail "stale-base included non-overlap: $stale"
) || exit 1

# sweep preserves the retry signal, and drain stops when review is ready.
(
  mkdir -p "$tmp/sweep-bin" "$tmp/sweep-state"
  cat > "$tmp/sweep-bin/gh" <<'GH'
#!/usr/bin/env bash
case "$*" in
  *'pr list'*) echo '[{"number":1,"headRefName":"feat/1-x"}]' ;;
  *mergeable*) echo MERGEABLE ;;
  *'pr checks'*) echo '[]' ;;
  *comments*) echo 1 ;;
  *) exit 0 ;;
esac
GH
  chmod +x "$tmp/sweep-bin/gh"
  printf '1\n' > "$tmp/sweep-state/reviewed.txt"
  # Stub the invoked land helper through a temporary copy of the script directory.
  cp -R "$SCRIPTS" "$tmp/sweep-scripts"
  cat > "$tmp/sweep-scripts/land.sh" <<'LAND'
#!/usr/bin/env bash
echo 'LANE-GATING slot=2 shared'
exit 4
LAND
  chmod +x "$tmp/sweep-scripts/land.sh"
  out=$(PATH="$tmp/sweep-bin:$PATH" BATCH_REPO="$ROOT" BATCH_STATE="$tmp/sweep-state" BATCH_BRANCH_RE='^feat/' bash "$tmp/sweep-scripts/sweep.sh") || fail "sweep failed: $out"
  grep -q '^LANE-GATING slot=2 shared$' <<<"$out" && ! grep -q '^FAIL' <<<"$out" || fail "sweep mishandled lane gating: $out"
  cat > "$tmp/sweep-scripts/land.sh" <<'LAND'
#!/usr/bin/env bash
echo 'WAIT: codecov/patch is still pending'
exit 3
LAND
  chmod +x "$tmp/sweep-scripts/land.sh"
  out=$(PATH="$tmp/sweep-bin:$PATH" BATCH_REPO="$ROOT" BATCH_STATE="$tmp/sweep-state" BATCH_BRANCH_RE='^feat/' bash "$tmp/sweep-scripts/sweep.sh") || fail "sweep failed on pending codecov: $out"
  grep -qx 'WAIT: codecov/patch is still pending' <<<"$out" || fail "sweep dropped codecov wait output: $out"
  ! grep -q '^FAIL' <<<"$out" || fail "sweep added FAIL for pending codecov: $out"
  cat > "$tmp/sweep-bin/sweep.sh" <<'SWEEP'
#!/usr/bin/env bash
echo 'REVIEW-READY 1'
SWEEP
  chmod +x "$tmp/sweep-bin/sweep.sh"
  # drain resolves its sibling sweep.sh from its script directory.
  # A private script directory allows the sweep stub and a fast gh response.
  mkdir -p "$tmp/drain-scripts"; cp "$SCRIPTS/common.sh" "$tmp/drain-scripts/common.sh"; cp "$SCRIPTS/drain.sh" "$tmp/drain-scripts/drain.sh"; cp "$tmp/sweep-bin/sweep.sh" "$tmp/drain-scripts/sweep.sh"
  rc=0; PATH="$tmp/sweep-bin:$PATH" BATCH_REPO="$ROOT" BATCH_STATE="$tmp/sweep-state" BATCH_BRANCH_RE='^feat/' timeout 20 bash "$tmp/drain-scripts/drain.sh" >"$tmp/drain.out" 2>&1 || rc=$?
  [[ "$rc" != 124 ]] || fail "drain timed out instead of stopping on REVIEW-READY: $(cat "$tmp/drain.out")"
  [[ "$rc" = 0 ]] || fail "drain did not exit 0 on REVIEW-READY: $(cat "$tmp/drain.out")"
) || exit 1

(
  setup_sync() {
    tag="$1"; SY_REPO="$tmp/sync-$tag-main"; SY_ROOT="$tmp/sync-$tag-lanes"; SY_STATE="$tmp/sync-$tag-state"
    mkdir -p "$SY_ROOT" "$SY_STATE"; git init -q --bare "$tmp/sync-$tag-origin.git"; git init -q "$SY_REPO"
    git -C "$SY_REPO" config user.email test@example.com; git -C "$SY_REPO" config user.name test
    printf 'base\n' > "$SY_REPO/tracked"; printf 'base\n' > "$SY_REPO/delete-me"; printf 'base\n' > "$SY_REPO/upstream"
    git -C "$SY_REPO" add tracked delete-me upstream; git -C "$SY_REPO" -c user.name=test -c user.email=test@example.com commit -qm base
    git -C "$SY_REPO" branch -M develop; git -C "$SY_REPO" remote add origin "$tmp/sync-$tag-origin.git"
    git -C "$SY_REPO" push -q -u origin develop; git -C "$SY_REPO" fetch -q origin
    git -C "$SY_REPO" worktree add -qb lane "$SY_ROOT/wt-ib1" origin/develop; SY_LANE="$SY_ROOT/wt-ib1"
  }
  setup_sync clean
  printf 'lane edit\n' > "$SY_LANE/tracked"; rm "$SY_LANE/delete-me"; printf new > "$SY_LANE/new-file"
  printf 'develop\n' >> "$SY_REPO/upstream"; git -C "$SY_REPO" add upstream; git -C "$SY_REPO" commit -qm upstream; git -C "$SY_REPO" push -q origin develop
  out=$(BATCH_REPO="$SY_REPO" BATCH_WT_ROOT="$SY_ROOT" BATCH_STATE="$SY_STATE" bash "$SCRIPTS/sync.sh" 1) || fail "sync clean failed: $out"
  grep -qx "backup: $SY_STATE/sync1.patch" <<<"$out" || fail 'sync backup missing'
  [ "$(git -C "$SY_LANE" rev-parse HEAD)" = "$(git -C "$SY_LANE" rev-parse origin/develop)" ] || fail 'sync HEAD incorrect'
  [ "$(cat "$SY_LANE/tracked")" = 'lane edit' ] && [ -f "$SY_LANE/new-file" ] && [ ! -e "$SY_LANE/delete-me" ] && grep -q develop "$SY_LANE/upstream" || fail 'sync did not preserve files'
  [ -f "$SY_STATE/sync1.patch" ] && [[ "$(git -C "$SY_LANE" status --porcelain)" == *'?? new-file'* ]] && [[ "$(git -C "$SY_LANE" status --porcelain)" == *' D delete-me'* ]] || fail 'sync backup/status incorrect'

  setup_sync conflict; printf 'lane\n' > "$SY_LANE/tracked"; printf 'keep me\n' > "$SY_LANE/conflict-untracked"; oldhead=$(git -C "$SY_LANE" rev-parse HEAD); oldstatus=$(git -C "$SY_LANE" status --porcelain)
  printf 'develop\n' > "$SY_REPO/tracked"; git -C "$SY_REPO" add tracked; git -C "$SY_REPO" commit -qm upstream; git -C "$SY_REPO" push -q origin develop
  if BATCH_REPO="$SY_REPO" BATCH_WT_ROOT="$SY_ROOT" BATCH_STATE="$SY_STATE" bash "$SCRIPTS/sync.sh" 1 >"$tmp/sync-conflict.out" 2>&1; then fail 'sync conflict succeeded'; else rc=$?; [ "$rc" = 2 ] || fail "conflict exit $rc"; fi
  grep -qx 'CONFLICT tracked' "$tmp/sync-conflict.out" || fail 'conflict filename missing'
  grep -qx 'lane kept on lane; to rebase by hand: git stash -u && git merge --ff-only origin/develop && git stash pop' "$tmp/sync-conflict.out" || fail 'conflict recovery instructions missing'
  [ "$(git -C "$SY_LANE" rev-parse HEAD)" = "$oldhead" ] && [ "$(git -C "$SY_LANE" status --porcelain)" = "$oldstatus" ] && [ "$(cat "$SY_LANE/tracked")" = lane ] && [ "$(cat "$SY_LANE/conflict-untracked")" = 'keep me' ] && [[ "$(git -C "$SY_LANE" status --porcelain)" == *'?? conflict-untracked'* ]] && ! git -C "$SY_LANE" rev-parse -q --verify MERGE_HEAD >/dev/null || fail 'conflict restore incorrect'

  setup_sync fetch-failure; printf 'local edit\n' > "$SY_LANE/tracked"; printf 'keep me\n' > "$SY_LANE/untracked"; oldhead=$(git -C "$SY_LANE" rev-parse HEAD); oldstatus=$(git -C "$SY_LANE" status --porcelain)
  git -C "$SY_LANE" remote set-url origin "$tmp/sync-fetch-failure-missing.git"
  if BATCH_REPO="$SY_REPO" BATCH_WT_ROOT="$SY_ROOT" BATCH_STATE="$SY_STATE" bash "$SCRIPTS/sync.sh" 1 >"$tmp/sync-fetch-failure.out" 2>&1; then fail 'sync with failed fetch succeeded'; else rc=$?; [ "$rc" = 1 ] || fail "fetch failure exit $rc"; fi
  [ "$(git -C "$SY_LANE" rev-parse HEAD)" = "$oldhead" ] && [ "$(git -C "$SY_LANE" status --porcelain)" = "$oldstatus" ] || fail 'fetch failure changed lane HEAD/status'

  setup_sync busy; printf dirty >> "$SY_LANE/tracked"; oldhead=$(git -C "$SY_LANE" rev-parse HEAD); oldstatus=$(git -C "$SY_LANE" status --porcelain)
  busy_scripts="$tmp/sync-busy-scripts"; mkdir -p "$busy_scripts"; cp "$SCRIPTS/common.sh" "$SCRIPTS/sync.sh" "$busy_scripts/"
  printf '#!/usr/bin/env bash\nexit 0\n' > "$busy_scripts/busy.sh"; chmod +x "$busy_scripts"/*.sh
  if BATCH_REPO="$SY_REPO" BATCH_WT_ROOT="$SY_ROOT" BATCH_STATE="$SY_STATE" bash "$busy_scripts/sync.sh" 1 >"$tmp/sync-busy.out" 2>&1; then fail 'busy sync succeeded'; else rc=$?; [ "$rc" = 1 ] || fail "busy exit $rc"; fi
  grep -qx 'REFUSED: a delegation is running in wt-ib1' "$tmp/sync-busy.out" || fail 'busy refusal missing'
  [ "$(git -C "$SY_LANE" rev-parse HEAD)" = "$oldhead" ] && [ "$(git -C "$SY_LANE" status --porcelain)" = "$oldstatus" ] || fail 'busy check changed lane'

  setup_sync commits; printf own >> "$SY_LANE/tracked"; git -C "$SY_LANE" add tracked; git -C "$SY_LANE" -c user.name=test -c user.email=test@example.com commit -qm own
  oldhead=$(git -C "$SY_LANE" rev-parse HEAD); oldstatus=$(git -C "$SY_LANE" status --porcelain)
  if BATCH_REPO="$SY_REPO" BATCH_WT_ROOT="$SY_ROOT" BATCH_STATE="$SY_STATE" bash "$SCRIPTS/sync.sh" 1 >"$tmp/sync-commits.out" 2>&1; then fail 'committed lane accepted'; else rc=$?; [ "$rc" = 1 ] || fail "committed exit $rc"; fi
  grep -qx 'REFUSED: lane has commits; merge by hand' "$tmp/sync-commits.out" || fail 'committed refusal missing'
  [ "$(git -C "$SY_LANE" rev-parse HEAD)" = "$oldhead" ] && [ "$(git -C "$SY_LANE" status --porcelain)" = "$oldstatus" ] || fail 'commit refusal changed lane'
) || exit 1
echo "batch-lanes script tests passed"
