#!/usr/bin/env bash
# PreToolUse(Bash): prevent process-pattern commands that can match their caller.
set -uo pipefail
payload=$(cat)
cmd=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""')
if ! command -v python3 >/dev/null 2>&1; then
  echo 'proc-kill-guard: cannot inspect command because python3 is unavailable' >&2
  echo '{}'; exit 0
fi
blocked=$(python3 - "$cmd" <<'PY'
import re, shlex, sys

def strip_heredocs(source):
    lines = source.splitlines(keepends=True)
    kept, i = [], 0
    marker = re.compile(r"<<(-?)\s*(?:'([^']+)'|\"([^\"]+)\"|([^\s;|&]+))")
    while i < len(lines):
        line = lines[i]; i += 1; kept.append(line)
        # Heredocs are consumed in declaration order; multiple declarations on
        # one command line are handled one after another.
        specs = [(bool(m.group(1)), next(x for x in m.groups()[1:] if x is not None)) for m in marker.finditer(line)]
        for strip_tabs, delimiter in specs:
            while i < len(lines):
                body = lines[i]; i += 1
                compare = body.lstrip('\t') if strip_tabs else body
                if compare.rstrip('\r\n') == delimiter: break
    return ''.join(kept)

def split(source):
    lexer = shlex.shlex(source, posix=True, punctuation_chars='|;&\n')
    lexer.whitespace = ' \t\r'
    lexer.whitespace_split = True
    return list(lexer)

def full_flag(opts):
    return not any(o == '-P' or o.startswith('-P') for o in opts) and any(o == '--full' or (o.startswith('-') and not o.startswith('--') and 'f' in o[1:]) for o in opts)

def leading_opts(words):
    opts = []
    for opt in words:
        if not opt.startswith('-'): break
        opts.append(opt)
    return opts

# Returns 'kill' for a kill that can hit its own shell, 'wait' for a wait loop
# that can never end because `pgrep -f` matches the loop's own command line,
# or '' when the command is fine.
def dangerous(source, depth=0):
    if depth > 1: return ''
    try: words = split(strip_heredocs(source))
    except ValueError: return ''
    segments, segment = [], []
    for word in words:
        if word and set(word) <= set('|;&\n'):
            if segment: segments.append(segment)
            segment = []
        else: segment.append(word)
    if segment: segments.append(segment)
    bad = False; pgrep_full = False; kill_sink = False; wait_loop = False
    loop_depth = 0
    for seg in segments:
        if not seg: continue
        head = seg[0].rsplit('/', 1)[-1]
        # Inside a while/until loop (its condition or its body), any `pgrep -f` is a wait
        # signal that also matches this shell: `while pgrep -f x`, `while [ -n "$(pgrep -f x)" ]`,
        # `while true; do pgrep -f x || break; done`.
        if head in ('while', 'until'): loop_depth += 1
        if loop_depth > 0:
            for k, word in enumerate(seg):
                if word.rsplit('/', 1)[-1] == 'pgrep' and full_flag(leading_opts(seg[k+1:])): wait_loop = True
                if '$(' in word and re.search(r'\$\(\s*(?:\S*/)?pgrep\s+(?:-\S*\s+)*(?:--full|-[A-Za-z]*f)', word): wait_loop = True
        if head == 'done' and loop_depth > 0: loop_depth -= 1
        if head in ('echo','printf','grep','rg','cat','sed','awk','head','tail','less','more'): continue
        for i, word in enumerate(seg):
            base = word.rsplit('/', 1)[-1]
            if base in ('bash','sh','dash','zsh') and '-c' in seg[i+1:]:
                ci = seg.index('-c', i+1)
                if ci+1 < len(seg):
                    inner = dangerous(seg[ci+1], depth+1)
                    if inner == 'kill': bad = True
                    elif inner == 'wait': wait_loop = True
            if base == 'pkill':
                for opt in seg[i+1:]:
                    if not opt.startswith('-') or opt == '--': break
                    if opt == '--full' or (opt.startswith('-') and not opt.startswith('--') and 'f' in opt[1:]): bad = True
            if base == 'killall' and '-r' in seg[i+1:]: bad = True
            if base == 'kill' or base.endswith('/kill'): kill_sink = True
            if base == 'xargs' and 'kill' in seg[i+1:]: kill_sink = True
            if base == 'pgrep' and full_flag(leading_opts(seg[i+1:])): pgrep_full=True
    if pgrep_full and kill_sink: bad=True
    joined=' '.join(words)
    if 'pgrep -f' in joined and ('kill $(' in joined or '| xargs kill' in joined or '| kill' in joined): bad=True
    if bad: return 'kill'
    return 'wait' if wait_loop else ''

print(dangerous(sys.argv[1]))
PY
)
if [ "$blocked" = wait ]; then
  jq -nc --arg msg 'Blocked: a while/until wait loop that runs `pgrep -f` (in its condition, in `$(…)`, or in its body) never ends, because `pgrep -f` also matches the shell running this loop (its command line contains the pattern). Wait on the work itself instead: poll for the line the job writes when it finishes (for example `until grep -q RC_DONE <log>; do sleep 5; done`), or for its PID (`while kill -0 <pid> 2>/dev/null; do sleep 5; done`).' '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:$msg}}'
elif [ "$blocked" = kill ]; then
  jq -nc --arg msg 'Blocked: `pkill -f` / `pgrep -f` and regex process matching can match the shell running this command and kill that shell too. Identify the target PID, then walk its children with `pgrep -P <pid>` and kill those PIDs (for example: `parent=<pid>; pgrep -P "$parent" | xargs -r kill`).' '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:$msg}}'
else
  echo '{}'
fi
