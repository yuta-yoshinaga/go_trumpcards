#!/usr/bin/env bash
set -uo pipefail
command -v jq >/dev/null 2>&1 || { echo '{}'; exit 0; }
payload=$(cat)
if ! printf '%s' "$payload" | jq -e . >/dev/null 2>&1; then echo '{}'; exit 0; fi
tool=$(printf '%s' "$payload" | jq -r '.tool_name // ""' 2>/dev/null) || { echo '{}'; exit 0; }
path=$(printf '%s' "$payload" | jq -r '.tool_input.file_path // ""' 2>/dev/null) || { echo '{}'; exit 0; }
[[ "$path" =~ \.test\.tsx?$ ]] || { echo '{}'; exit 0; }
text=''
case "$tool" in
  Edit) text=$(printf '%s' "$payload" | jq -r '.tool_input.new_string // ""' 2>/dev/null) || { echo '{}'; exit 0; } ;;
  MultiEdit) text=$(printf '%s' "$payload" | jq -r '[.tool_input.edits[]?.new_string // ""] | join("\n")' 2>/dev/null) || { echo '{}'; exit 0; } ;;
  Write) text=$(printf '%s' "$payload" | jq -r '.tool_input.content // ""' 2>/dev/null) || { echo '{}'; exit 0; } ;;
  *) echo '{}'; exit 0 ;;
esac
if printf '%s' "$text" | perl -0777 -ne 'print 1 if /(?<!\.not\.)toHaveTextContent\(\s*(?:\x27\x27|\"\")\s*\)/' | grep -q 1; then
  jq -nc --arg reason "toHaveTextContent('') matches any content; use toBeEmptyDOMElement() or assert real expected text." '{decision:"block",reason:$reason}'
else echo '{}'; fi
