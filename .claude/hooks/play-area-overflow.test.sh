#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/../commands/scripts" && pwd)
tmp=$(mktemp -d)
cleanup() { rm -rf "$tmp"; }
trap cleanup EXIT
mkdir -p "$tmp/bin"
cat > "$tmp/bin/agent-browser" <<'MOCK'
#!/usr/bin/env bash
printf '%s %s\n' "${AGENT_BROWSER_IDLE_TIMEOUT_MS-unset}" "$*" >> "$MOCK_LOG"
[[ ${MOCK_FAIL_OPEN:-0} == 1 && " $* " == *" open "* ]] && exit 1
if [[ ${MOCK_HANG:-0} == 1 && " $* " == *" wait "* ]]; then echo $$ > "$MOCK_LOG.hang"; exec sleep 30; fi
if [[ " $* " == *" eval "* ]]; then
  if [[ $* == *'document.documentElement.scrollWidth'* ]]; then echo 375
  elif [[ $* == *'game-footer-content'* && $* == *'overflow-blackjack-'* ]]; then echo '"nofooter"'
  elif [[ $* == *'game-footer-content'* && ${MOCK_BAD_FOOTER:-0} == 1 ]]; then echo garbage
  elif [[ $* == *'game-footer-content'* ]]; then echo 123
  elif [[ $* == *'overflow-blackjack-'* ]]; then echo '"nopa"'
  elif [[ $* == *'overflow-minibridge-'* && ${MOCK_BAD_VALUE:-0} == 1 ]]; then echo garbage
  else echo 123
  fi
fi
MOCK
chmod +x "$tmp/bin/agent-browser"
export MOCK_LOG="$tmp/log"

output=$(PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 -j 1 blackjack minibridge 2>"$tmp/stderr")
[[ $(printf '%s\n' "$output" | wc -l) -eq 3 ]]
[[ $(printf '%s\n' "$output" | sed -n '1p') == $'game\tmobile_worst\tmobile_footer\tdesktop\tscrollWidth' ]]
[[ $(printf '%s\n' "$output" | sed -n '2p') == $'blackjack\tnopa\tnofooter\tnopa\t375' ]]
[[ $(printf '%s\n' "$output" | sed -n '3p') == $'minibridge\t123\t123\t123\t375' ]]
grep -q 'mobile>300: 0, mobile_footer>40: 1, desktop>600: 0, scrollWidth>375: 0' "$tmp/stderr"
blackjack_close=$(grep -n 'overflow-blackjack-.* close$' "$tmp/log" | head -1 | cut -d: -f1)
minibridge_first=$(grep -n 'overflow-minibridge-.* set viewport' "$tmp/log" | head -1 | cut -d: -f1)
[[ -n $blackjack_close && -n $minibridge_first && $blackjack_close -lt $minibridge_first ]]
# The EXIT cleanup must not close sessions again: `close` on an unknown session starts a
# new daemon, so each session gets exactly one close (the worker's).
[[ $(grep -c 'overflow-blackjack-.* close$' "$tmp/log") -eq 1 ]]
[[ $(grep -c 'overflow-minibridge-.* close$' "$tmp/log") -eq 1 ]]

echo 'play-area-overflow: OK (serial game output and summary)'
: > "$tmp/log"
env -u AGENT_BROWSER_IDLE_TIMEOUT_MS PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -n 1 -j 1 blackjack >/dev/null 2>/dev/null
awk '$1 != 120000 { bad=1 } END { exit bad || NR == 0 }' "$tmp/log"
: > "$tmp/log"
AGENT_BROWSER_IDLE_TIMEOUT_MS=5000 PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -n 1 -j 1 blackjack >/dev/null 2>/dev/null
awk '$1 != 5000 { bad=1 } END { exit bad || NR == 0 }' "$tmp/log"

set +e
PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -j 5 blackjack >/dev/null 2>"$tmp/limit.err"
status=$?
set -e
[[ $status -eq 2 ]]
grep -q 'sessions must be between 1 and 4' "$tmp/limit.err"
PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -j 4 -n 1 blackjack >/dev/null 2>/dev/null

# A failed game must close its own session before the next game starts; the EXIT
# cleanup alone would also send a close, so only the ordering proves the worker did it.
: > "$tmp/log"
set +e
MOCK_BAD_VALUE=1 PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -n 1 -j 1 minibridge blackjack >/dev/null 2>/dev/null
status=$?
set -e
[[ $status -ne 0 ]]
minibridge_close=$(grep -n 'overflow-minibridge-.* close$' "$tmp/log" | head -1 | cut -d: -f1)
blackjack_first=$(grep -n 'overflow-blackjack-.* set viewport' "$tmp/log" | head -1 | cut -d: -f1)
[[ -n $minibridge_close && -n $blackjack_first && $minibridge_close -lt $blackjack_first ]]

# Invalid footer measurements fail the worker just like invalid play-area values.
set +e
MOCK_BAD_FOOTER=1 PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -n 1 -j 1 minibridge >/dev/null 2>/dev/null
status=$?
set -e
[[ $status -ne 0 ]]

# A failing browser command (not only an invalid measurement) must still fail the game.
set +e
MOCK_FAIL_OPEN=1 PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -n 1 -j 1 blackjack >/dev/null 2>/dev/null
status=$?
set -e
[[ $status -ne 0 ]]
echo 'play-area-overflow: OK (session cleanup, idle timeout, limits, failure cleanup)'

# SIGTERM mid-game: the script exits promptly with 143, stops the worker, and closes the
# session the worker had open.
: > "$tmp/log"
MOCK_HANG=1 PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -n 1 -j 1 blackjack >/dev/null 2>&1 &
script_pid=$!
for ((i = 0; i < 100; i++)); do [[ -s $tmp/log.hang ]] && break; sleep 0.1; done
[[ -s $tmp/log.hang ]]
kill -TERM "$script_pid"
set +e
wait "$script_pid"
status=$?
set -e
hung=$(cat "$tmp/log.hang")
for ((i = 0; i < 20; i++)); do kill -0 "$hung" 2>/dev/null || break; sleep 0.1; done
if kill -0 "$hung" 2>/dev/null; then kill "$hung"; echo 'worker tree survived SIGTERM' >&2; exit 1; fi
[[ $status -eq 143 ]]
grep -q 'overflow-blackjack-.* close$' "$tmp/log"
echo 'play-area-overflow: OK (no duplicate close, SIGTERM cleanup)'
