#!/usr/bin/env bash
# propose-games: idempotent, rate-limit-safe GitHub issue creation from a merged
# proposals JSON (array of {game,title,body}).
#
# Usage:
#   SRC=/tmp/newgames/all.json FOOTER='新規ゲーム追加提案バッチにより自動起票。' \
#     bash create_issues.sh
# Run in the background; ~3s/issue. Rerun to resume (skips already-created games).
set -uo pipefail

REPO_DIR="${REPO_DIR:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
SRC="${SRC:-/tmp/newgames/all.json}"
[[ "$SRC" = /* ]] || SRC="$PWD/$SRC"   # absolutize before the cd below
FOOTER="${FOOTER:-自動起票。実装時は docs/new-game-checklist.md に従うこと。}"
SLEEP="${SLEEP:-3}"                    # seconds between creates (rate-limit safety)
WORKDIR="$(dirname "$SRC")"
LOG="$WORKDIR/created.log"
ERR="$WORKDIR/errors.log"
BODY="$WORKDIR/body.md"

cd "$REPO_DIR"
mkdir -p "$WORKDIR"
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
  jq -r ".[$i].body" "$SRC" > "$BODY" || { echo "[$((i+1))/$n] $game -> jq extract failed, skip" >&2; continue; }
  printf '\n\n---\n_%s（スラッグ案: `%s`）_\n' "$FOOTER" "$game" >> "$BODY"
  if url=$(gh issue create --title "$title" --body-file "$BODY" 2>>"$ERR"); then
    printf '%s\t%s\t%s\n' "$game" "$url" "$title" >> "$LOG"
    echo "[$((i+1))/$n] $game -> $url"
  else
    echo "[$((i+1))/$n] $game -> FAILED (see $ERR)"
  fi
  sleep "$SLEEP"
done
echo "DONE. created=$(wc -l < "$LOG")"
