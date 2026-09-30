#!/usr/bin/env bash
set -uo pipefail
export LC_ALL=C.UTF-8
command -v jq >/dev/null 2>&1 || { echo '{}'; exit 0; }
payload=$(cat)
printf '%s' "$payload" | jq -e . >/dev/null 2>&1 || { echo '{}'; exit 0; }
cmd=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""' 2>/dev/null) || { echo '{}'; exit 0; }
if ! printf '%s' "$cmd" | grep -Eq '(^|[;&|[:space:]])gh[[:space:]]+pr[[:space:]]+(create|edit|comment)([[:space:]]|$)'; then echo '{}'; exit 0; fi
body=''
extract_heredoc() {
  local text=$1 delimiter strip line found=0 i start=0
  local heredoc_re="<<(-?)[[:space:]]*[\'\"]?([A-Za-z_][A-Za-z0-9_]*)[\'\"]?"
  if [[ "$text" =~ $heredoc_re ]]; then
    strip=${BASH_REMATCH[1]}; delimiter=${BASH_REMATCH[2]}
  else return 1; fi
  local -a lines=() body_lines=()
  mapfile -t lines <<< "$text"
  for ((i=0; i<${#lines[@]}; i++)); do
    if [[ ${lines[i]} == *'<<'* ]]; then start=$((i+1)); break; fi
  done
  for ((i=start; i<${#lines[@]}; i++)); do
    line=${lines[i]}
    [[ $strip == - ]] && line=${line#"$(printf '\t')"}
    if [[ $line == "$delimiter" ]]; then found=1; break; fi
    body_lines+=("${lines[i]}")
  done
  (( found )) || return 1
  body=$(printf '%s\n' "${body_lines[@]}")
}
if [[ "$cmd" =~ --body-file[[:space:]]+([^[:space:]]+) ]]; then
  path=${BASH_REMATCH[1]}; path=${path#\"}; path=${path%\"}; path=${path#\'}; path=${path%\'}
  if [ "$path" = '-' ]; then extract_heredoc "$cmd" || { echo '{}'; exit 0; }; else
    [ -r "$path" ] || { echo '{}'; exit 0; }
    body=$(cat -- "$path" 2>/dev/null) || { echo '{}'; exit 0; }
  fi
elif [[ "$cmd" =~ --body[[:space:]]+\"([^\"]*)\" ]]; then
  body=${BASH_REMATCH[1]}
  [[ "$body" == *'<<'* ]] && extract_heredoc "$body" || true
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
    if ! git grep --untracked -F -q -- "$segment" 2>/dev/null && { [ -z "$base" ] || ! git diff "$base" --unified=0 2>/dev/null | grep '^+' | grep -F -q -- "$segment"; }; then unverified+=("$segment"); fi
  done <<< "$fragment"
done < <(printf '%s\n' "$body" | grep -oE '`[^`]+`|「[^」]+」' | sed -E 's/^`|`$|^「|」$//g')
if [ ${#unverified[@]} -gt 0 ]; then
  joined=$(printf '• %s\n' "${unverified[@]}")
  jq -nc --arg msg "$joined" '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:("Unverified Japanese quote(s):\n" + $msg + "\nQuote the real string from the test/locale/diff, or add the line claim-check: skip to the body when the fragment is intentionally not in the repo (for example, screen output assembled at runtime).")}}'
else echo '{}'; fi
