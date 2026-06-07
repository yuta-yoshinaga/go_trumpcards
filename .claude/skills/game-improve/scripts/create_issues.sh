#!/usr/bin/env bash
# game-improve: idempotent, rate-limit-safe GitHub issue creation from a merged
# proposals JSON (array of {game,title,body}).
#
# Usage:
#   SRC=/tmp/claude-proposals/all.json REPO_DIR=/path/to/repo \
#     bash create_issues.sh
# Run in the background; ~3s/issue. Rerun to resume (skips already-created games).
set -uo pipefail

REPO_DIR="${REPO_DIR:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
SRC="${SRC:-/tmp/claude-proposals/all.json}"
[[ "$SRC" = /* ]] || SRC="$PWD/$SRC"   # absolutize before the cd below
WORKDIR="$(dirname "$SRC")"
LOG="$WORKDIR/created.log"
ERR="$WORKDIR/errors.log"
BODY="$WORKDIR/body.md"
SLEEP="${SLEEP:-3}"          # seconds between creates (secondary-rate-limit safety)

cd "$REPO_DIR"
touch "$LOG" "$ERR"

command -v jq >/dev/null || { echo "jq required" >&2; exit 1; }
command -v gh >/dev/null || { echo "gh CLI required" >&2; exit 1; }
gh auth status >/dev/null 2>&1 || { echo "gh not authenticated" >&2; exit 1; }
jq -e . "$SRC" >/dev/null 2>&1 || { echo "invalid JSON: $SRC" >&2; exit 1; }

n=$(jq 'length' "$SRC")
echo "Creating $n issues from $SRC (sleep ${SLEEP}s)..."
for ((i=0; i<n; i++)); do
  game=$(jq -r ".[$i].game" "$SRC")
  if grep -q "^${game}"$'\t' "$LOG" 2>/dev/null; then
    echo "[$((i+1))/$n] $game — already created, skip"
    continue
  fi
  title=$(jq -r ".[$i].title" "$SRC")
  jq -r ".[$i].body" "$SRC" > "$BODY"
  printf '\n\n---\n_対象ゲーム: `%s`。ゲーム改善提案バッチにより自動起票。_\n' "$game" >> "$BODY"
  url=$(gh issue create --title "$title" --body-file "$BODY" 2>>"$ERR")
  if [ -n "$url" ]; then
    printf '%s\t%s\t%s\n' "$game" "$url" "$title" >> "$LOG"
    echo "[$((i+1))/$n] $game -> $url"
  else
    echo "[$((i+1))/$n] $game -> FAILED (see $ERR)"
  fi
  sleep "$SLEEP"
done
echo "DONE. created=$(wc -l < "$LOG")"
