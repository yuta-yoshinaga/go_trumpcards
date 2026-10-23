#!/usr/bin/env bash
# Shared path configuration for batch-lanes scripts.
B="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="${BATCH_REPO:-$(git -C "$B/../../../../" rev-parse --show-toplevel)}"
REPO="$(cd "$REPO" && pwd)"
COMMON_DIR="$(git -C "$REPO" rev-parse --git-common-dir)"
[[ "$COMMON_DIR" = /* ]] || COMMON_DIR="$REPO/$COMMON_DIR"
BATCH_STATE="${BATCH_STATE:-$COMMON_DIR/batch-lanes}"
BATCH_WT_ROOT="${BATCH_WT_ROOT:-$(dirname "$REPO")}"
WT_ROOT="$BATCH_WT_ROOT"
mkdir -p "$BATCH_STATE"
touch "$BATCH_STATE/reviewed.txt" "$BATCH_STATE/acked.txt"
ALLOWLIST_SCRIPT="${B%/}/allowlist.sh"
if [[ ! -f "$ALLOWLIST_SCRIPT" ]]; then
  ALLOWLIST_SCRIPT="$PWD/.claude/skills/batch-lanes/scripts/allowlist.sh"
fi
source "$ALLOWLIST_SCRIPT"

# Match consecutive branch words against a game page name, allowing a CamelCase suffix.
batch_branch_matches_game() {
  local branch="$1" page="$2" compact game candidate start end
  local -a words=()
  IFS=- read -r -a words <<<"${branch##*/}"
  compact="$(tr '[:upper:]' '[:lower:]' <<<"$page" | tr -cd '[:alnum:]')"
  for ((start=0; start<${#words[@]}; start++)); do
    game=''
    for ((end=start; end<${#words[@]}; end++)); do
      game+="$(tr '[:upper:]' '[:lower:]' <<<"${words[end]}" | tr -cd '[:alnum:]')"
      [[ -n "$game" && "$compact" == "$game" ]] && return 0
      if [[ -n "$game" && "$compact" == "$game"* ]]; then
        candidate="${page:${#game}:1}"
        [[ -n "$candidate" && "$candidate" =~ [A-Z] ]] && return 0
      fi
    done
  done
  return 1
}

batch_branch_has_game() {
  local repo="$1" branch="$2" page
  while IFS= read -r page; do
    batch_branch_matches_game "$branch" "$page" && return 0
  done < <(sed -nE "s/.*page:[[:space:]]*['\"]([^'\"]+)['\"].*/\1/p" "$repo/frontend/src/constants/gameRoutes.ts")
  return 1
}

require_branch_re() {
  if [[ -z "${BATCH_BRANCH_RE:-}" ]]; then
    echo "Set BATCH_BRANCH_RE to select batch branches (for example: ^feat/[0-9]+-)" >&2
    return 2
  fi
}

rcg_live() {
  local slot="$1" pid="${2:-}"
  [[ -n "$pid" ]] || pid="$(cat "$BATCH_STATE/rcg$slot.pid" 2>/dev/null || true)"
  [[ "$pid" =~ ^[0-9]+$ ]] && [[ "$pid" != "$$" ]] && kill -0 "$pid" 2>/dev/null \
    && tr '\0' '\n' < "/proc/$pid/cmdline" 2>/dev/null | grep -Eq '(^|/)rcg[.]sh$' \
    && tr '\0' '\n' < "/proc/$pid/cmdline" 2>/dev/null | grep -Fxq -- "$slot"
}

collect() {
  local wt="$1"
  {
    git -C "$wt" diff --name-only -z origin/develop...HEAD 2>/dev/null || true
    git -C "$wt" diff --name-only -z HEAD 2>/dev/null || true
    git -C "$wt" ls-files --others --exclude-standard -z 2>/dev/null || true
  } | sort -zu
}

stale_base_files() {
  local wt="$1" base
  base="$(git -C "$wt" merge-base HEAD origin/develop 2>/dev/null)" || return 0
  comm -12 <(collect "$wt" | tr '\0' '\n' | LC_ALL=C sort -u) \
    <(git -C "$wt" diff --name-only "$base" origin/develop 2>/dev/null | LC_ALL=C sort -u)
}

delegated_to() {
  local slot="$1" log backend branch_file current_branch
  if [[ -n "${BATCH_DELEGATED_TO:-}" ]]; then printf '%s' "$BATCH_DELEGATED_TO"; return; fi
  branch_file="$BATCH_STATE/slot$slot.branch"
  current_branch=$(git -C "$WT_ROOT/wt-ib$slot" branch --show-current 2>/dev/null || true)
  if [[ ! -f "$branch_file" || -z "$current_branch" || "$(cat "$branch_file")" != "$current_branch" ]]; then
    printf 'unknown'
    return
  fi
  local -a backends=()
  shopt -s nullglob
  for log in "$BATCH_STATE"/go-"$slot"-*.log; do
    [[ "$log" -nt "$branch_file" ]] || continue
    while IFS= read -r backend; do [[ -n "$backend" ]] && backends+=("$backend"); done \
      < <(sed -nE 's/.*dele: ([^[:space:]]+) ok.*/\1/p' "$log")
  done
  shopt -u nullglob
  if ((${#backends[@]} == 0)); then printf 'none (orchestrator)'; else
    local -a unique=()
    mapfile -t unique < <(printf '%s\n' "${backends[@]}" | LC_ALL=C sort -u)
    local joined='' item
    for item in "${unique[@]}"; do joined+="${joined:+, }$item"; done
    printf '%s' "$joined"
  fi
}
