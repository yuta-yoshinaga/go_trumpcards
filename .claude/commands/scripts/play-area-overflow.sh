#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: play-area-overflow.sh [-p port] [-n deals] [-j sessions] [-o out.tsv] [game ...]

Measure games on an already running go_trumpcards web server. Defaults: port 8080,
3 mobile reloads per game, 4 parallel browser sessions, and stdout TSV output.
EOF
}

port=8080 deals=3 sessions=4 output=-
while getopts ':p:n:j:o:h' opt; do
  case "$opt" in
    p) port=$OPTARG ;;
    n) deals=$OPTARG ;;
    j) sessions=$OPTARG ;;
    o) output=$OPTARG ;;
    h) usage; exit 0 ;;
    :) echo "missing value for -$OPTARG" >&2; usage >&2; exit 2 ;;
    \?) echo "unknown option: -$OPTARG" >&2; usage >&2; exit 2 ;;
  esac
done
shift $((OPTIND - 1))
[[ $port =~ ^[0-9]+$ && $deals =~ ^[1-9][0-9]*$ && $sessions =~ ^[1-9][0-9]*$ ]] || { echo 'port must be numeric; deals and sessions must be positive integers' >&2; exit 2; }
command -v agent-browser >/dev/null || { echo 'agent-browser is required' >&2; exit 127; }

routes=frontend/src/constants/gameRoutes.ts
root_game=$(grep -A1 -P "path:\s*'/'," "$routes" | grep -oP "labelKey:\s*'nav\.\K[^']*")
if (($#)); then games=("$@"); else
  mapfile -t games < <(grep -oP "path:\s*'/\K[^']*" "$routes" | sed "s/^$/$root_game/")
fi
((${#games[@]})) || { echo 'no game routes found' >&2; exit 1; }
mapfile -t all_games < <(grep -oP "path:\s*'/\K[^']*" "$routes" | sed "s/^$/$root_game/")

tmp=$(mktemp -d)
cleanup() {
  for game in "${games[@]}"; do agent-browser --session "overflow-$game-$$" close >/dev/null 2>&1 || true; done
  for file in "$tmp"/*; do [[ -e $file ]] && unlink "$file"; done
  rmdir "$tmp" 2>/dev/null || true
}
trap cleanup EXIT
base="http://localhost:$port"
keys=$(printf '%s\n' "${all_games[@]}" "${games[@]}" | sort -u | sed 's/.*/tutorial_completed_&/; s/.*/"&": "true"/' | paste -sd, -)
init="for (const [k,v] of Object.entries({$keys})) localStorage.setItem(k,v)"

printf 'game\tmobile_worst\tdesktop\tscrollWidth\n' > "$tmp/results"
eval_value() {
  local value
  value=$(agent-browser "$@" | tail -1)
  value=$(printf '%s' "$value" | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')
  if [[ $value == \"*\" && ${#value} -ge 2 ]]; then value=${value:1:${#value}-2}; fi
  printf '%s\n' "$value"
}
worker() {
  local game=$1 session="overflow-$1-$$" path url mobile=nopa desktop width value i
  path="/#/$game"; [[ $game == "$root_game" ]] && path=/
  url="$base$path"
  agent-browser --session "$session" set viewport 375 667 >/dev/null
  agent-browser --session "$session" open "$url" >/dev/null
  agent-browser --session "$session" eval "$init" >/dev/null
  for ((i=0; i<deals; i++)); do
    agent-browser --session "$session" open "$url" >/dev/null
    agent-browser --session "$session" reload >/dev/null
    agent-browser --session "$session" wait '[aria-live="polite"]' >/dev/null
    agent-browser --session "$session" wait 900 >/dev/null
    value=$(eval_value --session "$session" eval "(()=>{const e=document.querySelector('main [aria-busy] > div.flex-1.overflow-y-auto');return e?e.scrollHeight-e.clientHeight:'nopa'})()")
    [[ $value == nopa || $value =~ ^[0-9]+$ ]] || { echo "$game: invalid mobile measurement: $value" >&2; return 1; }
    if [[ $value =~ ^[0-9]+$ ]] && { [[ $mobile == nopa ]] || (( value > mobile )); }; then mobile=$value; fi
  done
  agent-browser --session "$session" set viewport 1280 800 >/dev/null
  agent-browser --session "$session" open "$url" >/dev/null
  agent-browser --session "$session" wait '[aria-live="polite"]' >/dev/null
  agent-browser --session "$session" wait 900 >/dev/null
  desktop=$(eval_value --session "$session" eval "(()=>{const e=document.querySelector('main [aria-busy] > div.flex-1.overflow-y-auto');return e?e.scrollHeight-e.clientHeight:'nopa'})()")
  agent-browser --session "$session" set viewport 375 667 >/dev/null
  width=$(eval_value --session "$session" eval 'document.documentElement.scrollWidth')
  [[ ( $desktop == nopa || $desktop =~ ^[0-9]+$ ) && $width =~ ^[0-9]+$ ]] || { echo "$game: invalid desktop/width measurement: $desktop/$width" >&2; return 1; }
  printf '%s\t%s\t%s\t%s\n' "$game" "$mobile" "$desktop" "$width" > "$tmp/$game.tsv"
}

active=0 failed=0
for game in "${games[@]}"; do
  worker "$game" &
  ((active+=1))
  if ((active >= sessions)); then wait -n || failed=1; active=$((active - 1)); fi
done
while ((active)); do wait -n || failed=1; active=$((active - 1)); done
((failed == 0)) || { echo 'one or more browser measurements failed' >&2; exit 1; }
for game in "${games[@]}"; do cat "$tmp/$game.tsv" >> "$tmp/results"; done
if [[ $output == - ]]; then cat "$tmp/results"; else cp "$tmp/results" "$output"; fi

awk -F '\t' 'NR>1 {if ($2 ~ /^[0-9]+$/ && $2>300) m++; if ($3 ~ /^[0-9]+$/ && $3>600) d++; if ($4 ~ /^[0-9]+$/ && $4>375) w++} END {printf "mobile>300: %d, desktop>600: %d, scrollWidth>375: %d\n", m, d, w}' "$tmp/results" >&2
