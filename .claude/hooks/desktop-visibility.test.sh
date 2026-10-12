#!/usr/bin/env bash
set -euo pipefail
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/../commands/scripts" && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin"
cat > "$tmp/bin/ss" <<'MOCKSS'
#!/usr/bin/env bash
printf 'LISTEN 0 128 *:19001 *:* users:(("fake",pid=%s,fd=3))\n' "${MOCK_SS_PID:?}"
MOCKSS
cat > "$tmp/bin/agent-browser" <<'MOCK'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "$MOCK_LOG"
if [[ " $* " == *' eval '* ]]; then
  if [[ $* == *"querySelector("* ]]; then echo 'nohand'
  elif [[ ${MOCK_ZERO:-0} == 1 ]]; then echo '0/0'
  else echo '3/2'; fi
fi
MOCK
chmod +x "$tmp/bin/ss" "$tmp/bin/agent-browser"
export MOCK_SS_PID=$$ MOCK_LOG="$tmp/log"

out=$(PATH="$tmp/bin:$PATH" bash "$script_dir/desktop-visibility.sh" -p 19001 blackjack solitaire 2>"$tmp/board.err")
[[ $(sed -n '1p' <<<"$out") == $'game\tfaceup\thidden' ]]
[[ $(sed -n '2p' <<<"$out") == $'blackjack\t3\t2' ]]
grep -q 'games with hidden > 0: 2' "$tmp/board.err"

MOCK_ZERO=1 PATH="$tmp/bin:$PATH" bash "$script_dir/desktop-visibility.sh" -p 19001 blackjack >/dev/null 2>"$tmp/zero.err"
grep -q 'warning: blackjack: faceup=0; a count of nothing is not a pass' "$tmp/zero.err"

out=$(PATH="$tmp/bin:$PATH" bash "$script_dir/desktop-visibility.sh" -p 19001 -m hand blackjack='.player-hand' 2>"$tmp/hand.err")
[[ $(sed -n '1p' <<<"$out") == $'game\tcards\thidden' ]]
[[ $(sed -n '2p' <<<"$out") == $'blackjack\tnohand\t0' ]]
grep -q 'warning: blackjack: cards=0; a count of nothing is not a pass' "$tmp/hand.err"

set +e
PATH="$tmp/bin:$PATH" bash "$script_dir/desktop-visibility.sh" -m invalid blackjack >/dev/null 2>"$tmp/usage.err"
status=$?
set -e
[[ $status -eq 2 ]]
echo 'desktop-visibility: OK'
