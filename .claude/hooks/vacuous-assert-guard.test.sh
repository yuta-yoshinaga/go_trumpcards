#!/usr/bin/env bash
set -uo pipefail
HOOK="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/vacuous-assert-guard.sh"
fail=0
run() {
  local expect=$1 tool=$2 path=$3 field=$4 text=$5 desc=$6 out blocked=pass
  out=$(jq -nc --arg tool "$tool" --arg path "$path" --arg field "$field" --arg text "$text" '{tool_name:$tool,tool_input:({file_path:$path} + (if $field == "edits" then {edits:[{new_string:$text}]} else {($field):$text} end))}' | bash "$HOOK")
  printf '%s' "$out" | grep -q '"decision":"block"' && blocked=block
  if [ "$blocked" = "$expect" ]; then echo "  ok    $desc ($expect)"; else echo "  FAIL  $desc expected $expect: $out"; fail=1; fi
}
echo 'vacuous-assert-guard:'
run block Edit a.test.ts new_string 'x.toHaveTextContent("")' 'Edit blocks'
run pass Edit a.test.ts new_string 'x.toHaveTextContent("ok")' 'Edit allows valid assertion'
run block MultiEdit a.test.tsx edits "x.toHaveTextContent('')" 'MultiEdit blocks'
run pass MultiEdit a.test.tsx edits 'toBeEmptyDOMElement()' 'MultiEdit allows valid assertion'
run block Write a.test.tsx content "x.toHaveTextContent('')" 'Write blocks'
run pass Write a.test.tsx content "x.toHaveTextContent('text')" 'Write allows valid assertion'
run pass Edit a.ts new_string "toHaveTextContent('')" 'non-test file'
run pass Edit a.test.ts new_string "x.not.toHaveTextContent('')" 'not form'
[ "$fail" -eq 0 ] || exit 1
echo 'vacuous-assert-guard: all cases passed'
