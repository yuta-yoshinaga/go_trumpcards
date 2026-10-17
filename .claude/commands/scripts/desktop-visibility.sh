#!/usr/bin/env bash
set -euo pipefail
export AGENT_BROWSER_IDLE_TIMEOUT_MS="${AGENT_BROWSER_IDLE_TIMEOUT_MS:-120000}"

usage() {
  cat <<'EOF'
Usage: desktop-visibility.sh [-p port] [-W width] [-H height] [-m board|hand] [game ... | game=selector ...]

Measure desktop card visibility on an already running go_trumpcards server. Defaults:
port 8080, viewport 1280x800, board mode. Hand mode requires each game as
<game>=<CSS selector for the hand container>. A card is hidden when fewer than
52px at its top remain visible. The server must belong to this checkout.
EOF
}

port=8080 width=1280 height=800 mode=board
while getopts ':p:W:H:m:h' opt; do
  case "$opt" in
    p) port=$OPTARG ;;
    W) width=$OPTARG ;;
    H) height=$OPTARG ;;
    m) mode=$OPTARG ;;
    h) usage; exit 0 ;;
    :) echo "missing value for -$OPTARG" >&2; usage >&2; exit 2 ;;
    \?) echo "unknown option: -$OPTARG" >&2; usage >&2; exit 2 ;;
  esac
done
shift $((OPTIND - 1))
[[ $port =~ ^[0-9]+$ && $width =~ ^[1-9][0-9]*$ && $height =~ ^[1-9][0-9]*$ ]] || { echo 'port must be numeric; width and height must be positive integers' >&2; exit 2; }
[[ $mode == board || $mode == hand ]] || { echo 'mode must be board or hand' >&2; usage >&2; exit 2; }
if [[ $mode == hand && $# -eq 0 ]]; then echo 'hand mode requires game=selector arguments' >&2; exit 2; fi
command -v agent-browser >/dev/null || { echo 'agent-browser is required' >&2; exit 127; }
command -v ss >/dev/null || { echo 'ss (iproute2) is required to identify the server' >&2; exit 127; }
top=$(git rev-parse --show-toplevel)
listening=$(ss -ltnpH "sport = :$port" 2>/dev/null || true)
[[ -n $listening ]] || { echo "no server is listening on port $port" >&2; exit 2; }
pid=$(sed -n 's/.*pid=\([0-9][0-9]*\).*/\1/p;T;q' <<<"$listening")
cwd=''
[[ -z $pid ]] || cwd=$(readlink "/proc/$pid/cwd" 2>/dev/null || true)
if [[ -z $pid || -z $cwd ]]; then echo "cannot identify the server process on port $port" >&2; exit 2; fi
if [[ $cwd != "$top" ]]; then echo "server on port $port (pid $pid) runs from $cwd, not $top" >&2; exit 2; fi
echo "server: pid $pid cwd $cwd" >&2

routes=frontend/src/constants/gameRoutes.ts
root_game=$(grep -A1 -P "path:\s*'/'," "$routes" | grep -oP "labelKey:\s*'nav\.\K[^']*")
if (($#)); then games=("$@"); elif [[ $mode == board ]]; then
  mapfile -t games < <(grep -oP "path:\s*'/\K[^']*" "$routes" | sed "s/^$/$root_game/")
else
  echo 'hand mode requires game=selector arguments' >&2; exit 2
fi
((${#games[@]})) || { echo 'no game routes found' >&2; exit 2; }
if [[ $mode == hand ]]; then
  for item in "${games[@]}"; do [[ $item == *=* && -n ${item#*=} ]] || { echo "invalid hand argument: $item (expected game=selector)" >&2; exit 2; }; done
fi

session="desktop-visibility-$$"
opened=0
cleanup() { local status=$?; trap - EXIT INT TERM HUP; if ((opened)); then agent-browser --session "$session" close >/dev/null 2>&1 || true; fi; exit "$status"; }
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
base="http://localhost:$port"
keys=$(printf '%s\n' "${games[@]}" | sed 's/=.*//' | sed 's/.*/tutorial_completed_&/' | sort -u | sed 's/.*/"&": "true"/' | paste -sd, -)
init="for (const [k,v] of Object.entries({$keys})) localStorage.setItem(k,v)"
agent-browser --session "$session" set viewport "$width" "$height" >/dev/null
opened=1
printf 'game\t%s\thidden\n' "$([[ $mode == board ]] && echo faceup || echo cards)"
hidden_games=0
for item in "${games[@]}"; do
  game=${item%%=*}; selector=${item#*=}
  [[ $mode == board ]] && game=$item
  path="/#/$game"; [[ $game == "$root_game" ]] && path=/
  agent-browser --session "$session" open "$base$path" >/dev/null
  agent-browser --session "$session" eval "$init" >/dev/null
  agent-browser --session "$session" open "$base$path" >/dev/null
  agent-browser --session "$session" wait '[aria-live="polite"]' >/dev/null
  agent-browser --session "$session" wait 300 >/dev/null
  agent-browser --session "$session" wait --fn "document.getAnimations().every(a => a.playState==='finished' || a.playState==='idle' || a.effect?.getComputedTiming().iterations===Infinity)" >/dev/null
  if [[ $mode == board ]]; then
    js='(()=>{const elems=[...document.querySelectorAll("img,[role=img]")];let face=0,hidden=0;for(const e of elems){const r=e.getBoundingClientRect(),label=((e.getAttribute("alt")||"")+" "+(e.getAttribute("aria-label")||""));if(r.width<20||r.height<30||/裏|back/i.test(label))continue;face++;let b=innerHeight;for(let a=e.parentElement;a;a=a.parentElement){const s=getComputedStyle(a);if(["auto","scroll","hidden"].includes(s.overflowY)&&a.scrollHeight>a.clientHeight)b=Math.min(b,a.getBoundingClientRect().bottom)}if(r.top+52>b)hidden++}return face+"/"+hidden})()'
    label=faceup
  else
    selector_json=$(printf '%s' "$selector" | jq -Rs .)
    js="(()=>{const c=document.querySelector($selector_json);if(!c)return 'nohand';let cards=0,hidden=0;for(const e of c.querySelectorAll('img,[role=img]')){const r=e.getBoundingClientRect(),label=(e.getAttribute('alt')||'')+' '+(e.getAttribute('aria-label')||'');if(r.width<20||r.height<30||/裏|back/i.test(label))continue;cards++;let b=innerHeight;for(let a=e.parentElement;a;a=a.parentElement){const s=getComputedStyle(a);if(['auto','scroll','hidden'].includes(s.overflowY)&&a.scrollHeight>a.clientHeight)b=Math.min(b,a.getBoundingClientRect().bottom)}if(r.top+52>b)hidden++}return cards+'/'+hidden})()"
    label=cards
  fi
  value=$(agent-browser --session "$session" eval "$js" | tail -1 | tr -d '"[:space:]')
  if [[ $value == nohand ]]; then count=nohand; hidden=0
  elif [[ $value =~ ^([0-9]+)/([0-9]+)$ ]]; then count=${BASH_REMATCH[1]}; hidden=${BASH_REMATCH[2]}
  else echo "$game: invalid $label/hidden measurement: $value" >&2; exit 1; fi
  if [[ $count == nohand || $count == 0 ]]; then echo "warning: $game: $label=0; a count of nothing is not a pass" >&2; fi
  ((hidden == 0)) || ((hidden_games+=1))
  printf '%s\t%s\t%s\n' "$game" "$count" "$hidden"
done
echo "games with hidden > 0: $hidden_games" >&2
