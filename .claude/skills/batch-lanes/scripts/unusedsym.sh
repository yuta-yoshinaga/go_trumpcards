#!/usr/bin/env bash
set -uo pipefail
worktree="${1:?usage: unusedsym.sh <worktree>}"
cd "$worktree" || exit 2
python3 - "$worktree" <<'PY'
import os, re, subprocess, sys
root = sys.argv[1]
allowed = set(os.environ.get("ALLOW_UNUSED", "").split())
changed = subprocess.check_output(["git", "diff", "--name-only", "-z", "HEAD"]).decode().split("\0")[:-1]
untracked = subprocess.check_output(["git", "ls-files", "--others", "--exclude-standard", "-z"]).decode().split("\0")[:-1]
files = dict.fromkeys(changed + untracked)
defs = []
go_patterns = [re.compile(r'^\s*func\s+(?:\([^)]*\)\s*)?([A-Za-z_][A-Za-z0-9_]*)\s*\(')]
ts_pattern = re.compile(r'^\s*export\s+(?:declare\s+)?(?:async\s+)?(?:function|class|const|let|var)\s+([A-Za-z_$][\w$]*)')
for path in files:
    if not path or not os.path.isfile(path): continue
    ext = os.path.splitext(path)[1]
    if ext not in ('.go', '.ts', '.tsx'): continue
    if path.endswith(('_test.go', '_mock.go')) or re.search(r'\.test\.tsx?$', path): continue
    if path in untracked:
        additions = [(n, line) for n, line in enumerate(open(path, encoding='utf-8', errors='replace'), 1)]
    else:
        diff = subprocess.check_output(['git', 'diff', '--no-ext-diff', '-U0', 'HEAD', '--', path], text=True, errors='replace')
        additions = []
        new_line = 0
        for line in diff.splitlines():
            m = re.match(r'@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@', line)
            if m: new_line = int(m.group(1)); continue
            if line.startswith('+') and not line.startswith('+++'): additions.append((new_line, line[1:])); new_line += 1
            elif not line.startswith('-'): new_line += 1
    for lineno, line in additions:
        match = (go_patterns[0].match(line) if ext == '.go' else ts_pattern.match(line))
        if match: defs.append((match.group(1), path, lineno))

for name, path, lineno in defs:
    if name in allowed: continue
    result = subprocess.run(['grep', '-R', '-n', '-I', '-w', '--exclude-dir=.git', '--exclude-dir=node_modules', '--exclude-dir=vendor', '--', name, '.'], text=True, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
    hits = [row for row in result.stdout.splitlines() if not (row.startswith(f'./{path}:') and row.split(':', 2)[1] == str(lineno))]
    if not hits: print(f'WARN UNUSED_NEW_SYMBOL {name} {path}')
PY
