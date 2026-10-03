#!/usr/bin/env bash
set -uo pipefail
HOOK="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/proc-kill-guard.sh"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
cd "$TMP" || exit 1
# The guard only inspects command text, so no repository fixture is needed.
fail=0
run() {
  local expect=$1 desc=$2 cmd=$3 out blocked=pass
  out=$(printf '{"tool_input":{"command":%s}}' "$(printf '%s' "$cmd" | jq -Rs .)" | bash "$HOOK")
  printf '%s' "$out" | grep -q '"permissionDecision": *"deny"' && blocked=block
  if [ "$blocked" = "$expect" ]; then echo "  ok $desc ($expect)"; else echo "  FAIL $desc expected=$expect got=$blocked $out"; fail=1; fi
  if [ "$blocked" = block ]; then
    printf '%s' "$out" | grep -q 'self\|shell\|pgrep -P' || { echo "  FAIL $desc missing explanation"; fail=1; }
    ! printf '%s' "$out" | grep -q '"continue" *: *false' || { echo "  FAIL $desc used continue false"; fail=1; }
  fi
}
echo 'proc-kill-guard:'
run block 'pkill full flag' 'pkill -f worker'
run block 'pkill long flag with options reordered' 'pkill -i --full worker'
run block 'combined pkill flags' 'pkill -fx worker'
run block 'pgrep output piped to kill' 'pgrep -f worker | xargs kill'
run block 'pgrep command substitution to kill' 'kill $(pgrep -f worker)'
run block 'killall regex' 'killall -r worker.*'
run block 'newline command separator' $'echo start\npkill -f worker'
run block 'bash -c command' "bash -c 'pkill -f x'"
run block 'sh -c command' 'sh -c "pkill -f x"'
run pass 'pkill exact' 'pkill -x worker'
run pass 'pkill plain pattern' 'pkill worker'
run pass 'kill PID' 'kill 1234'
run pass 'pgrep parent result to kill' 'kill $(pgrep -P 1234)'
run pass 'pgrep display only' 'pgrep -f worker'
run pass 'grep phrase' "grep 'pkill -f' file"
run pass 'echo phrase' 'echo "pkill -f"'
run pass 'heredoc phrase' $'cat <<EOF\npkill -f worker\nEOF'
run pass 'quoted heredoc phrase' $"git commit -F - <<'EOF'\npkill -f worker\nEOF"
run pass 'tab stripped heredoc' $'cat <<-EOF\n\tpkill -f worker\n\tEOF'
run pass 'ssh command is out of scope' "ssh host 'pkill -f x'"
run pass 'xargs command string is out of scope' "printf '%s' 'pkill -f x' | xargs sh -c"
[ "$fail" -eq 0 ] || exit 1
echo 'proc-kill-guard: all cases passed'
