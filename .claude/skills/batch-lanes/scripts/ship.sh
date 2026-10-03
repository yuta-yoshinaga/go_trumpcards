#!/bin/bash
source "$(dirname "${BASH_SOURCE[0]}")/common.sh"
# refuse while a delegation is still writing in this worktree (2026-10-01: a re-run shipped #9574's half-written files as #9564)
$B/busy.sh $1 >/dev/null && { echo "REFUSED: a delegation is still running in wt-ib$1"; exit 1; }
# refuse unless rcg.sh passed on exactly this worktree state
[ "$(cat "$BATCH_STATE/gate$1.ok" 2>/dev/null)" = "$("$B/treehash.sh" "$1")" ] || { echo "REFUSED: no passing rcg.sh for slot $1 at this worktree state"; exit 1; }
# usage: ship.sh <slot> <issue#> "<commit subject>" "<summary lines (markdown)>"
set -e
s=$1 n=$2 subj=$3 summ=$4
cd "$WT_ROOT/wt-ib$s"
br=$(git branch --show-current)
files=$(git status --porcelain | awk '$2!="-"{print $2}')
git add $files
git commit -q -F - <<M
$subj

$summ

Closes #$n

Delegated-To: codex (dele -k edit)
Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
${CLAUDE_SESSION_URL:+Claude-Session: $CLAUDE_SESSION_URL}
M
git push -q -u origin $br 2>&1 | grep -v '^remote:' || true
tn=$(git show HEAD -U0 -- "*.test.ts" "*.test.tsx" "*_test.go" | grep -E "^\+\s*(it|test|describe)\(|^\+func Test" | sed -E "s/^\+\s*//; s/, (async )?\(\) => \{$//; s/ \{$//" | head -12)
[ -n "$tn" ] && summ="$summ

**Tests added/changed (from the diff):**
$(sed "s/^/- /" <<<"$tn")"
tp_fe=''; tp_go=''; BT='`'
git show --stat HEAD --format= | grep -qE '\.(tsx?|mjs)' && tp_fe="- [x] ${BT}bunx vitest run${BT} on the changed test files and their shared-component siblings"
git show --stat HEAD --format= | grep -qE '\.go|internal/i18n' && tp_go="- [x] ${BT}go build ./...${BT}, ${BT}go test -tags test ./...${BT}, ${BT}golangci-lint run --build-tags test ./internal/...${BT}"
body=$(cat <<M
## Summary
$summ

Closes #$n

## Test plan
$tp_fe
- [x] \`bun run check\` / \`bun run typecheck\`
$tp_go
- [ ] CI green

Delegated-To: codex (dele -k edit)

${CLAUDE_SESSION_URL:+🤖 Generated with [Claude Code](https://claude.com/claude-code)

$CLAUDE_SESSION_URL}
M
)
gh pr create --base develop --head $br --title "$subj" --body "$body"
