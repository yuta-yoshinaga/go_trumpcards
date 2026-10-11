#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/../commands/scripts" && pwd)
tmp=$(mktemp -d)
foreign_pid=''
cleanup() {
  [[ -n ${foreign_pid:-} ]] && kill "$foreign_pid" 2>/dev/null || true
  rm -rf "$tmp"
}
trap cleanup EXIT
mkdir -p "$tmp/bin"
cat > "$tmp/bin/ss" <<'MOCKSS'
#!/usr/bin/env bash
case ${MOCK_SS_MODE:-same} in
  none) exit 0 ;;
  unknown) printf 'LISTEN 0 128 *:19001 *:* users:(("fake",pid=999999999,fd=3))\n' ;;
  foreign) printf 'LISTEN 0 128 *:19001 *:* users:(("fake",pid=%s,fd=3))\n' "${MOCK_FOREIGN_PID:?}" ;;
  *) printf 'LISTEN 0 128 *:19001 *:* users:(("fake",pid=%s,fd=3))\n' "${MOCK_SS_PID:?}" ;;
esac
MOCKSS
cat > "$tmp/bin/agent-browser" <<'MOCK'
#!/usr/bin/env bash
printf '%s %s\n' "${AGENT_BROWSER_IDLE_TIMEOUT_MS-unset}" "$*" >> "$MOCK_LOG"
[[ ${MOCK_FAIL_OPEN:-0} == 1 && " $* " == *" open "* ]] && exit 1
if [[ ${MOCK_HANG:-0} == 1 && " $* " == *" wait "* ]]; then echo $$ > "$MOCK_LOG.hang"; exec sleep 30; fi
if [[ " $* " == *" eval "* ]]; then
  if [[ $* == *'document.documentElement.scrollWidth'* ]]; then echo 375
  elif [[ $* == *'game-footer-content'* && $* == *'overflow-blackjack-'* ]]; then echo '"nofooter"'
  elif [[ $* == *'52>b'* && ${MOCK_BAD_CARDS:-0} == 1 ]]; then echo garbage
  elif [[ $* == *'52>b'* && ${MOCK_CHANGING_CARDS:-0} == 1 ]]; then
    count=$(cat "$MOCK_LOG.count" 2>/dev/null || echo 0); count=$((count + 1)); echo "$count" > "$MOCK_LOG.count"
    if ((count == 1)); then echo '1/30'; else echo '3/20'; fi
  elif [[ $* == *'52>b'* && ${MOCK_ZERO_CARDS:-0} == 1 ]]; then echo '0/0'
  elif [[ $* == *'52>b'* && ${MOCK_CARDS:-0} == 1 ]]; then echo '2/34'
  elif [[ $* == *'52>b'* ]]; then echo '0/3'
  elif [[ $* == *'game-footer-content'* && ${MOCK_BAD_FOOTER:-0} == 1 ]]; then echo garbage
  elif [[ $* == *'game-footer-content'* ]]; then echo 123
  elif [[ $* == *'overflow-blackjack-'* ]]; then echo '"nopa"'
  elif [[ $* == *'overflow-minibridge-'* && ${MOCK_BAD_VALUE:-0} == 1 ]]; then echo garbage
  else echo 123
  fi
fi
MOCK
chmod +x "$tmp/bin/agent-browser" "$tmp/bin/ss"
export MOCK_LOG="$tmp/log"
export MOCK_SS_PID=$$

output=$(PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 -j 1 blackjack minibridge 2>"$tmp/stderr")
[[ $(printf '%s\n' "$output" | wc -l) -eq 3 ]]
[[ $(printf '%s\n' "$output" | sed -n '1p') == $'game\tmobile_worst\tmobile_footer\tdesktop\tscrollWidth\thidden_cards\tfooter_cards' ]]
[[ $(printf '%s\n' "$output" | sed -n '2p') == $'blackjack\tnopa\tnofooter\tnopa\t375\tnofooter\tnofooter' ]]
[[ $(printf '%s\n' "$output" | sed -n '3p') == $'minibridge\t123\t123\t123\t375\t0\t3' ]]
grep -q 'mobile>300: 0, mobile_footer>40: 1, desktop>600: 0, scrollWidth>375: 0, hidden_cards>0: 0, footer_cards=0: 0' "$tmp/stderr"
blackjack_close=$(grep -n 'overflow-blackjack-.* close$' "$tmp/log" | head -1 | cut -d: -f1)
minibridge_first=$(grep -n 'overflow-minibridge-.* set viewport' "$tmp/log" | head -1 | cut -d: -f1)
[[ -n $blackjack_close && -n $minibridge_first && $blackjack_close -lt $minibridge_first ]]
# The EXIT cleanup must not close sessions again: `close` on an unknown session starts a
# new daemon, so each session gets exactly one close (the worker's).
[[ $(grep -c 'overflow-blackjack-.* close$' "$tmp/log") -eq 1 ]]
[[ $(grep -c 'overflow-minibridge-.* close$' "$tmp/log") -eq 1 ]]

# Card counts: alternate mock, reload aggregation, zero-count warning, and invalid value.
: > "$MOCK_LOG"
output=$(MOCK_CARDS=1 PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 -j 1 minibridge 2>"$tmp/cards.err")
[[ $(printf '%s\n' "$output" | sed -n '2p') == $'minibridge\t123\t123\t123\t375\t2\t34' ]]
: > "$MOCK_LOG"
output=$(MOCK_CHANGING_CARDS=1 PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 2 -j 1 minibridge 2>"$tmp/reloads.err")
[[ $(printf '%s\n' "$output" | sed -n '2p') == $'minibridge\t123\t123\t123\t375\t3\t20' ]]
# Both reloads must wait for animations before querying hidden/cards.
for reload in 1 2; do
  wait_line=$(grep -n 'wait --fn' "$MOCK_LOG" | sed -n "${reload}p" | cut -d: -f1)
  cards_line=$(grep -n '52>b' "$MOCK_LOG" | sed -n "${reload}p" | cut -d: -f1)
  [[ -n $wait_line && -n $cards_line && $wait_line -lt $cards_line ]]
done
: > "$MOCK_LOG"
output=$(MOCK_ZERO_CARDS=1 PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 -j 1 minibridge 2>"$tmp/zero.err")
[[ $(printf '%s\n' "$output" | sed -n '2p' | awk -F '\t' '{print $6 "\t" $7}') == $'0\t0' ]]
grep -q 'warning: minibridge: no cards counted in game-footer-content; hidden_cards=0 checks nothing' "$tmp/zero.err"
grep -q 'footer_cards=0: 1' "$tmp/zero.err"
set +e
MOCK_BAD_CARDS=1 PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 -j 1 minibridge >/dev/null 2>"$tmp/bad-cards.err"
status=$?
set -e
[[ $status -ne 0 ]]
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

# Server ownership checks: no listener and a foreign cwd, with and without -x.
set +e
MOCK_SS_MODE=none PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 blackjack >/dev/null 2>"$tmp/no-listener.err"
status=$?
set -e
[[ $status -eq 2 ]]
grep -q 'no server is listening on port 19001' "$tmp/no-listener.err"
(cd "$tmp" && sleep 30) &
foreign_pid=$!
mkdir -p "$tmp/public"
: > "$tmp/public/index.html"
export MOCK_FOREIGN_PID=$foreign_pid
set +e
MOCK_SS_MODE=foreign PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 blackjack >/dev/null 2>"$tmp/foreign.err"
status=$?
set -e
[[ $status -eq 2 ]]
grep -q 'runs from' "$tmp/foreign.err"
MOCK_SS_MODE=foreign PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 -x blackjack >/dev/null 2>"$tmp/foreign-x.err"
grep -q 'server: pid' "$tmp/foreign-x.err"
kill "$foreign_pid" 2>/dev/null || true
wait "$foreign_pid" 2>/dev/null || true
foreign_pid=''

# An unknown server PID fails closed without -x and warns while continuing with -x.
set +e
MOCK_SS_MODE=unknown PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 blackjack >/dev/null 2>"$tmp/unknown.err"
status=$?
set -e
[[ $status -eq 2 ]]
grep -q 'cannot identify the server process on port 19001' "$tmp/unknown.err"
MOCK_SS_MODE=unknown PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 -x blackjack >/dev/null 2>"$tmp/unknown-x.err"
grep -q 'warning: cannot identify the server process on port 19001' "$tmp/unknown-x.err"

# Missing ss is reported as a dependency error before querying listeners.
set +e
mkdir -p "$tmp/no-ss-bin"
ln -s "$tmp/bin/agent-browser" "$tmp/no-ss-bin/agent-browser"
PATH="$tmp/no-ss-bin" /bin/bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 blackjack >/dev/null 2>"$tmp/missing-ss.err"
status=$?
set -e
[[ $status -eq 127 ]]
grep -q 'ss (iproute2) is required to identify the server' "$tmp/missing-ss.err"

# An external server cwd without public/index.html gets the actionable build warning.
mkdir -p "$tmp/missing"
(cd "$tmp/missing" && sleep 30) &
foreign_pid=$!
export MOCK_FOREIGN_PID=$foreign_pid
MOCK_SS_MODE=foreign PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 -x blackjack >/dev/null 2>"$tmp/missing-index.err"
grep -q "warning: $tmp/missing/public/index.html does not exist; build with (cd frontend && bun run build)" "$tmp/missing-index.err"
kill "$foreign_pid" 2>/dev/null || true
wait "$foreign_pid" 2>/dev/null || true
foreign_pid=''
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
