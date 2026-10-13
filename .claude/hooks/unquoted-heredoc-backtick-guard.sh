#!/usr/bin/env bash
# PreToolUse(Bash): block backticks in unquoted heredoc bodies.
set -uo pipefail
command -v jq >/dev/null 2>&1 || { echo '{}'; exit 0; }
payload=$(cat)
cmd=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""')
command -v python3 >/dev/null 2>&1 || { echo '{}'; exit 0; }
blocked=$(python3 - "$cmd" <<'PY'
import re
import sys

source = sys.argv[1]
# A delimiter is quoted when its word uses single quotes, double quotes, or
# a backslash escape. The negative lookbehind excludes here-strings (<<<).
decl = re.compile(r"(?<!<)<<(-?)\s*(?:'([^']+)'|\"([^\"]+)\"|\\([^\s;|&]+)|([^\s;|&]+))")
lines = source.splitlines()
consumed = set()
for i, line in enumerate(lines):
    if i in consumed:
        continue
    specs = []
    for m in decl.finditer(line):
        quoted = any(m.group(n) is not None for n in (2, 3, 4))
        delimiter = next(m.group(n) for n in (2, 3, 4, 5) if m.group(n) is not None)
        specs.append((m.group(1) == '-', delimiter, quoted))
    body_line = i + 1
    for strip_tabs, delimiter, quoted in specs:
        # Only a terminated heredoc counts: `$((1<<2))` or `# see <<EOF` never
        # reaches a line equal to its "delimiter", so it is not a heredoc at all.
        body, start = [], body_line
        while body_line < len(lines):
            compare = lines[body_line].lstrip('\t') if strip_tabs else lines[body_line]
            if compare.rstrip('\r') == delimiter:
                break
            body.append(lines[body_line])
            body_line += 1
        else:
            body_line = start
            continue
        consumed.update(range(start, body_line + 1))
        body_line += 1
        if not quoted and any('`' in b for b in body):
            print('deny')
            raise SystemExit
print('')
PY
)
if [ "$blocked" = deny ]; then
  jq -nc --arg msg "Blocked: backticks in an unquoted heredoc body are expanded by the shell. Quote the delimiter, for example use <<'EOF'." '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:$msg}}'
else
  echo '{}'
fi
