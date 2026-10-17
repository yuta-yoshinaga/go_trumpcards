#!/usr/bin/env bash
# PreToolUse(Bash): require narrowly defined docs when staged registrations add
# an ADR, Web API route, or game.
set -uo pipefail

payload=$(cat)
cmd=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""')
case "$cmd" in
  git\ commit*) ;;
  *) echo '{}'; exit 0 ;;
esac

# DOC_DRIFT_RANGE is for deterministic history backtests. Normal hook calls
# inspect the index, including the staged blob rather than the working tree.
if [ -n "${DOC_DRIFT_RANGE:-}" ]; then
  base_ref=${DOC_DRIFT_RANGE%%..*}
  name_status=$(git diff --name-status "$DOC_DRIFT_RANGE")
  added_lines=$(git diff --no-ext-diff --unified=0 "$DOC_DRIFT_RANGE" | awk '/^\+\+\+ / { file=substr($0,7); next } /^\+/ && !/^\+\+\+/ { print file "\t" substr($0,2) }')
else
  base_ref=HEAD
  name_status=$(git diff --cached --name-status)
  added_lines=$(git diff --cached --no-ext-diff --unified=0 | awk '/^\+\+\+ / { file=substr($0,7); next } /^\+/ && !/^\+\+\+/ { print file "\t" substr($0,2) }')
fi

staged_paths=$(printf '%s\n' "$name_status" | awk 'NF { if ($1 ~ /^R/) { print $3 } else { print $2 } }')
reason=''
has_path() { printf '%s\n' "$staged_paths" | grep -Fxq -- "$1"; }

# R1: only a newly added ADR file requires a README index entry.
if printf '%s\n' "$name_status" | awk '$1 == "A" && $2 ~ /^docs\/adr\/[0-9][0-9][0-9][0-9]-.+\.md$/ { found=1 } END { exit !found }'; then
  has_path docs/adr/README.md || reason+=$'\n- docs/adr/README.md (new numbered ADRs must be indexed)'
fi

# R2: route registrations live in these files today. Match only literal
# ServeMux registrations with an HTTP path; the SPA fallback is not a route.
new_route=$(printf '%s\n' "$added_lines" | grep -E '^internal/infrastructure/web/[^[:space:]]+[[:space:]].*HandleFunc\("(GET |POST |PUT |PATCH |DELETE )?/[^" ]+"' | head -1)
if [ -n "$new_route" ] && ! has_path api/openapi.yaml; then
  reason+=$'\n- api/openapi.yaml (a Web API route registration was added)'
fi

# R3: only inspect one-line `{Name: ..., Category: ...}` literals; multiline
# entries are intentionally out of scope. Adding a registry literal changes the canonical game list. Existing
# entry edits do not match this added-line-only check.
new_game=0
while IFS= read -r line; do
  case "$line" in *'internal/infrastructure/games/registry.go'*'{Name:'*'Category:'*) ;;
    *) continue ;;
  esac
  game=$(printf '%s\n' "$line" | sed -nE 's/.*Name:[[:space:]]*"([^"]+)".*/\1/p')
  [ -n "$game" ] || continue
  if ! git show "$base_ref:internal/infrastructure/games/registry.go" 2>/dev/null | grep -Fq "Name: \"$game\""; then
    new_game=1
    break
  fi
done <<< "$added_lines"
if [ "$new_game" -eq 1 ]; then
  has_path README.md || reason+=$'\n- README.md (the canonical game table must be updated)'
  has_path docs/games.md || reason+=$'\n- docs/games.md (the game catalog must be updated)'
  printf '%s\n' "$staged_paths" | grep -Eq '^docs/manual/' || reason+=$'\n- docs/manual/ (a manual page must be staged)'
fi

if [ -n "$reason" ]; then
  jq -nc --arg reason "${reason#$'\n'}" '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:("Documentation missing for staged additions:\n" + $reason)}}'
else
  echo '{}'
fi
exit 0
