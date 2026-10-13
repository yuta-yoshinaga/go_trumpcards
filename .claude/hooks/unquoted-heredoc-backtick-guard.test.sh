#!/usr/bin/env bash
set -uo pipefail
HOOK="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/unquoted-heredoc-backtick-guard.sh"
command -v jq >/dev/null 2>&1 || { echo 'skip jq is not installed'; exit 0; }
fail=0
run() {
  local expect=$1 desc=$2 cmd=$3 out blocked=pass
  out=$(printf '{"tool_input":{"command":%s}}' "$(printf '%s' "$cmd" | jq -Rs .)" | bash "$HOOK")
  printf '%s' "$out" | grep -q '"permissionDecision": *"deny"' && blocked=block
  if [ "$blocked" = "$expect" ]; then echo "  ok $desc ($expect)"; else echo "  FAIL $desc expected=$expect got=$blocked $out"; fail=1; fi
  if [ "$blocked" = block ]; then
    printf '%s' "$out" | grep -q "<<'EOF'" || { echo "  FAIL $desc missing quoted delimiter guidance"; fail=1; }
  fi
}
echo 'unquoted-heredoc-backtick-guard:'
run block 'unquoted delimiter' $'cat <<EOF\n`foo`\nEOF'
run block 'space before delimiter' $'cat << EOF\n`foo`\nEOF'
run block 'tab-stripping delimiter' $'cat <<-EOF\n\t`foo`\n\tEOF'
run block 'multiple heredocs second unsafe' $'cat <<A <<B\nplain\nA\n`unsafe`\nB'
run block 'multiple heredocs first unsafe' $'cat <<A <<B\n`unsafe`\nA\nplain\nB'
run pass 'backtick outside body' $'echo `date`; cat <<EOF\nplain\nEOF'
run pass 'single quoted delimiter' $'cat <<\'EOF\'\n`foo`\nEOF'
run pass 'double quoted delimiter' $'cat <<"EOF"\n`foo`\nEOF'
run pass 'escaped delimiter' $'cat <<\\EOF\n`foo`\nEOF'
run pass 'dollar command substitution is allowed' $'cat <<EOF\n$(date)\nEOF'
run pass 'variable expansion is allowed' $'cat <<EOF\n$HOME\nEOF'
run pass 'here-string is outside scope' 'cat <<< `foo`'
run pass 'arithmetic shift is not a heredoc' $'echo $((1<<2))\necho `date`'
run pass 'unterminated <<WORD in a comment is not a heredoc' $'x=1 # see <<EOF\necho `date`'
[ "$fail" -eq 0 ] || exit 1
echo 'unquoted-heredoc-backtick-guard: all cases passed'
