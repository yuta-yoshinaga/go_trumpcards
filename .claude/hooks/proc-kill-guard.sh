#!/usr/bin/env bash
# PreToolUse(Bash): prevent process-pattern commands that can match their caller.
set -uo pipefail
payload=$(cat)
cmd=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""')

# Parse shell words without evaluating command substitutions. Literal text in
# echo/grep and heredocs is not an executable process command.
blocked=$(python3 - "$cmd" <<'PY'
import shlex, sys
source = sys.argv[1]
try:
    lexer = shlex.shlex(source, posix=True, punctuation_chars='|;&')
    lexer.whitespace_split = True
    words = list(lexer)
except ValueError:
    print('')
    raise SystemExit
segments, segment = [], []
for word in words:
    if word and set(word) <= set('|;&'):
        if segment: segments.append(segment)
        segment = []
    else:
        segment.append(word)
if segment: segments.append(segment)
bad = False
pgrep_full = False
kill_sink = False
for seg in segments:
    if not seg: continue
    head = seg[0].rsplit('/', 1)[-1]
    if head in ('echo', 'printf', 'grep', 'rg', 'cat', 'sed', 'awk', 'head', 'tail', 'less', 'more'):
        continue
    for i, word in enumerate(seg):
        base = word.rsplit('/', 1)[-1]
        if base == 'pkill':
            for opt in seg[i+1:]:
                if not opt.startswith('-'): break
                if opt == '--': break
                if opt == '--full' or (opt.startswith('-') and not opt.startswith('--') and 'f' in opt[1:]): bad = True
        if base == 'killall' and '-r' in seg[i+1:]: bad = True
        if base == 'kill' or base.endswith('/kill'): kill_sink = True
        if base == 'xargs' and 'kill' in seg[i+1:]: kill_sink = True
        if base == 'pgrep':
            opts = []
            for opt in seg[i+1:]:
                if not opt.startswith('-'): break
                opts.append(opt)
            if not any(o == '-P' or o.startswith('-P') for o in opts) and any(o == '--full' or (o.startswith('-') and not o.startswith('--') and 'f' in o[1:]) for o in opts):
                pgrep_full = True
if pgrep_full and kill_sink: bad = True
joined = ' '.join(words)
if 'pgrep -f' in joined and ('kill $(' in joined or '| xargs kill' in joined or '| kill' in joined): bad = True
print('yes' if bad else '')
PY
)
if [ "$blocked" = yes ]; then
  jq -nc --arg msg 'Blocked: `pkill -f` / `pgrep -f` and regex process matching can match the shell running this command and kill that shell too. Identify the target PID, then walk its children with `pgrep -P <pid>` and kill those PIDs (for example: `parent=<pid>; pgrep -P "$parent" | xargs -r kill`).' '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:$msg}}'
else
  echo '{}'
fi
