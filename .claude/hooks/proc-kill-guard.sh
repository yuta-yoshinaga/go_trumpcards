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

def dangerous(source, depth=0):
    if depth > 1: return False
    try: words = split(strip_heredocs(source))
    except ValueError: return False
    segments, segment = [], []
    for word in words:
        if word and set(word) <= set('|;&\n'):
            if segment: segments.append(segment)
            segment = []
        else: segment.append(word)
    if segment: segments.append(segment)
    bad = False; pgrep_full = False; kill_sink = False
    for seg in segments:
        if not seg: continue
        head = seg[0].rsplit('/', 1)[-1]
        if head in ('echo','printf','grep','rg','cat','sed','awk','head','tail','less','more'): continue
        for i, word in enumerate(seg):
            base = word.rsplit('/', 1)[-1]
            if base in ('bash','sh','dash','zsh') and '-c' in seg[i+1:]:
                ci = seg.index('-c', i+1)
                if ci+1 < len(seg) and dangerous(seg[ci+1], depth+1): bad = True
            if base == 'pkill':
                for opt in seg[i+1:]:
                    if not opt.startswith('-') or opt == '--': break
                    if opt == '--full' or (opt.startswith('-') and not opt.startswith('--') and 'f' in opt[1:]): bad = True
            if base == 'killall' and '-r' in seg[i+1:]: bad = True
            if base == 'kill' or base.endswith('/kill'): kill_sink = True
            if base == 'xargs' and 'kill' in seg[i+1:]: kill_sink = True
            if base == 'pgrep':
                opts=[]
                for opt in seg[i+1:]:
                    if not opt.startswith('-'): break
                    opts.append(opt)
                if not any(o == '-P' or o.startswith('-P') for o in opts) and any(o == '--full' or (o.startswith('-') and not o.startswith('--') and 'f' in o[1:]) for o in opts): pgrep_full=True
    if pgrep_full and kill_sink: bad=True
    joined=' '.join(words)
    if 'pgrep -f' in joined and ('kill $(' in joined or '| xargs kill' in joined or '| kill' in joined): bad=True
    return bad

print('yes' if dangerous(sys.argv[1]) else '')
PY
)
if [ "$blocked" = yes ]; then
  jq -nc --arg msg 'Blocked: `pkill -f` / `pgrep -f` and regex process matching can match the shell running this command and kill that shell too. Identify the target PID, then walk its children with `pgrep -P <pid>` and kill those PIDs (for example: `parent=<pid>; pgrep -P "$parent" | xargs -r kill`).' '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:$msg}}'
else
  echo '{}'
fi
