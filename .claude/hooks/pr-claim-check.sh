#!/usr/bin/env bash
set -uo pipefail
export LC_ALL=C.UTF-8
command -v jq >/dev/null 2>&1 || { echo '{}'; exit 0; }
payload=$(cat)
printf '%s' "$payload" | jq -e . >/dev/null 2>&1 || { echo '{}'; exit 0; }
cmd=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""' 2>/dev/null) || { echo '{}'; exit 0; }
if ! printf '%s' "$cmd" | grep -Eq '(^|[;&|[:space:]])gh[[:space:]]+pr[[:space:]]+(create|edit|comment)([[:space:]]|$)'; then echo '{}'; exit 0; fi
body=''
if [[ "$cmd" =~ --body-file[[:space:]]+([^[:space:]]+) ]]; then
  path=${BASH_REMATCH[1]}; path=${path#\"}; path=${path%\"}; path=${path#\'}; path=${path%\'}
  [ "$path" = '-' ] && { echo '{}'; exit 0; }
  [ -r "$path" ] || { echo '{}'; exit 0; }
  body=$(cat -- "$path" 2>/dev/null) || { echo '{}'; exit 0; }
elif [[ "$cmd" =~ --body[[:space:]]+\"([^\"]*)\" ]]; then body=${BASH_REMATCH[1]}
elif [[ "$cmd" =~ --body[[:space:]]+\'([^\']*)\' ]]; then body=${BASH_REMATCH[1]}
else echo '{}'; exit 0
fi
[[ "$body" == *'claim-check: skip'* ]] && { echo '{}'; exit 0; }
unverified=()
while IFS= read -r fragment; do
  printf '%s' "$fragment" | grep -Pq '[\p{Hiragana}\p{Katakana}\p{Han}]' || continue
  # Split (newline) on placeholders, digits and ellipses: each piece must exist verbatim.
  fragment=$(printf '%s' "$fragment" | sed -E 's/\{\{[^}]*\}\}/\n/g; s/(^|[^[:alpha:]])[NMX]([^[:alpha:]]|$)/\1\n\2/g; s/[0-9]+([.][0-9]+)?/\n/g; s/…|\.\.\./\n/g')
  while IFS= read -r segment; do
    segment=$(printf '%s' "$segment" | sed -E 's/^[[:space:]]+|[[:space:]]+$//g')
    printf '%s' "$segment" | grep -Pq '[\p{Hiragana}\p{Katakana}\p{Han}]' || continue
    count=$(printf '%s' "$segment" | perl -CSD -ne '$n = () = /[\p{Hiragana}\p{Katakana}\p{Han}]/g; print $n')
    [ "$count" -ge 3 ] || continue
    base=$(git merge-base origin/develop HEAD 2>/dev/null) || base=''
    if ! git grep -F -q -- "$segment" 2>/dev/null && { [ -z "$base" ] || ! git diff "$base" --unified=0 2>/dev/null | grep '^+' | grep -F -q -- "$segment"; }; then unverified+=("$segment"); fi
  done <<< "$fragment"
done < <(printf '%s\n' "$body" | grep -oE '`[^`]+`|「[^」]+」' | sed -E 's/^`|`$|^「|」$//g')
if [ ${#unverified[@]} -gt 0 ]; then
  joined=$(printf '• %s\n' "${unverified[@]}")
  jq -nc --arg msg "$joined" '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:("Unverified Japanese quote(s):\n" + $msg + "\nQuote the real string from the test/locale/diff, or add the line claim-check: skip to the body when the fragment is intentionally not in the repo (for example, screen output assembled at runtime).")}}'
else echo '{}'; fi
