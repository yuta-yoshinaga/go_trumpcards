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
if [[ "${ALLOW_BRANCH_ISSUE_MISMATCH:-}" != 1 && "$br" =~ /([0-9]+)- && "${BASH_REMATCH[1]}" != "$n" ]]; then
  echo "REFUSED: branch $br issue ${BASH_REMATCH[1]} does not match issue $n" >&2
  exit 1
fi
if [[ "${ALLOW_UNMENTIONED_ISSUE:-}" != 1 ]] && grep -Eq '#[0-9]+' <<<"$summ" && ! grep -Eq "#$n([^0-9]|$)" <<<"$summ"; then
  echo "REFUSED: PR body mentions another issue but not #$n" >&2
  exit 1
fi
while IFS= read -r line; do
  while IFS= read -r closing; do
    closed=${closing##*#}
    [[ "$closed" == "$n" ]] || echo "WARN: body closes #$closed" >&2
  done < <(grep -Eio '\b(close[sd]?|fix(e[sd])?|resolve[sd]?)\s+#[0-9]+' <<<"$line" || true)
done <<<"$summ"
mapfile -d '' -t files < <({ git diff --name-only -z HEAD; git ls-files -o --exclude-standard -z; } | sort -zu)
if [[ "${ALLOW_STALE_BASE:-}" != 1 ]]; then
  if ! git fetch -q origin develop; then
    echo "REFUSED: could not fetch origin/develop to verify this lane's base" >&2
    exit 1
  fi
  base=$(git merge-base HEAD origin/develop) || { echo "REFUSED: could not find merge base with origin/develop" >&2; exit 1; }
  upstream=$(git diff --name-only "$base" origin/develop)
  overlap=$(comm -12 <(printf '%s\n' "${files[@]}" | LC_ALL=C sort -u) <(printf '%s\n' "$upstream" | LC_ALL=C sort -u))
  if [[ -n "$overlap" ]]; then
    printf '%s\n' "$overlap" >&2
    echo "REFUSED: develop changed these files after this lane branched; merge origin/develop into the lane first" >&2
    exit 1
  fi
fi
if ((${#files[@]})); then git add -- "${files[@]}"; fi
metadata=()
# BATCH_DELEGATED_TO remains the explicit override for inferred lane authors.
delegation=$(delegated_to "$s")
metadata+=("Delegated-To: $delegation")
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
