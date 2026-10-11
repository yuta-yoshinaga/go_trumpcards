#!/usr/bin/env bash
set -euo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
tmp=$(mktemp -d)
cleanup() { rm -rf "$tmp"; }
trap cleanup EXIT
mkdir -p "$tmp/bin"
cat > "$tmp/bin/agent-browser" <<'MOCK'
#!/usr/bin/env bash
if [[ " $* " == *" eval "* ]]; then
  if [[ $* == *'document.documentElement.scrollWidth'* ]]; then echo 375; elif [[ $* == *overflow-blackjack-* ]]; then echo '"nopa"'; else echo 123; fi
fi
MOCK
chmod +x "$tmp/bin/agent-browser"

output=$(PATH="$tmp/bin:$PATH" bash "$script_dir/play-area-overflow.sh" -p 19001 -n 1 -j 2 blackjack minibridge 2>"$tmp/stderr")
[[ $(printf '%s\n' "$output" | wc -l) -eq 3 ]]
[[ $(printf '%s\n' "$output" | sed -n '1p') == $'game\tmobile_worst\tdesktop\tscrollWidth' ]]
[[ $(printf '%s\n' "$output" | sed -n '2p') == $'blackjack\tnopa\tnopa\t375' ]]
[[ $(printf '%s\n' "$output" | sed -n '3p') == $'minibridge\t123\t123\t375' ]]
grep -q 'mobile>300: 0, desktop>600: 0, scrollWidth>375: 0' "$tmp/stderr"
echo 'play-area-overflow: OK (parallel game output and summary)'
