#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
source "$(dirname "${BASH_SOURCE[0]}")/prbody.sh"
# refuse while a delegation is still writing in this worktree (2026-10-01: a re-run shipped #9574's half-written files as #9564)
$B/busy.sh "$1" >/dev/null && { echo "REFUSED: a delegation is still running in wt-ib$1"; exit 1; }
# refuse unless rcg.sh passed on exactly this worktree state
[ "$(cat "$BATCH_STATE/gate$1.ok" 2>/dev/null)" = "$("$B/treehash.sh" "$1")" ] || { echo "REFUSED: no passing rcg.sh for slot $1 at this worktree state"; exit 1; }
# usage: ship.sh <slot> <issue#> "<commit subject>" "<summary lines (markdown)> | <full PR body with ## headings>"
set -e
s="$1" n="$2" subj="$3" summ="$4"
cd "$WT_ROOT/wt-ib$s"
if [[ "${CLAIM_SKIP:-}" != 1 ]] && ! "$B/claimcheck.sh" "$PWD" "$summ"; then
  echo "REFUSED: unverified claims" >&2
  exit 1
fi
br=$(git branch --show-current)
mapfile -d '' -t files < <({ git diff --name-only -z HEAD; git ls-files -o --exclude-standard -z; } | sort -zu)
if ((${#files[@]})); then git add -- "${files[@]}"; fi
metadata=()
[[ -n "${BATCH_DELEGATED_TO:-codex (dele -k edit)}" ]] && metadata+=("Delegated-To: ${BATCH_DELEGATED_TO:-codex (dele -k edit)}")
[[ -n "${CLAUDE_CO_AUTHOR:-}" ]] && metadata+=("Co-Authored-By: $CLAUDE_CO_AUTHOR")
closes_line="Closes #$n"
prbody_has_ref "$n" "$summ" && closes_line=''
git commit -q -F - <<M
$subj

$summ

$closes_line

${metadata[*]}
${CLAUDE_SESSION_URL:+Claude-Session: $CLAUDE_SESSION_URL}
M
set +e
git push -q -u origin "$br" 2>&1 | grep -v '^remote:'
push_status=${PIPESTATUS[0]}
set -e
((push_status == 0)) || exit "$push_status"
tn=$(git show HEAD -U0 -- "*.test.ts" "*.test.tsx" "*_test.go" | grep -E "^\+\s*(it|test|describe)\(|^\+func Test" | sed -E "s/^\+\s*//; s/, (async )?\(\) => \{$//; s/ \{$//" | head -12)
[ -n "$tn" ] && ! prbody_is_full "$summ" && summ="$summ

**Tests added/changed (from the diff):**
$(sed "s/^/- /" <<<"$tn")"
tp_fe=''; tp_go=''; BT='`'
git show --stat HEAD --format= | grep -qE '\.(tsx?|mjs)' && tp_fe="- [x] ${BT}bunx vitest run${BT} on the changed test files and their shared-component siblings"
git show --stat HEAD --format= | grep -qE '\.go|internal/i18n' && tp_go="- [x] ${BT}go build ./...${BT}, ${BT}go test -tags test ./...${BT}, ${BT}golangci-lint run --build-tags test ./internal/...${BT}"
body=$(ship_pr_body "$n" "$summ" "$tp_fe" "$tp_go")
gh pr create --base develop --head $br --title "$subj" --body "$body"
