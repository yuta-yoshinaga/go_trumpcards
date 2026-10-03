#!/usr/bin/env bash
# PreToolUse(Bash): protect branches checked out in another worktree.
set -uo pipefail
payload=$(cat)
cmd=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""')
if ! command -v python3 >/dev/null 2>&1; then
  echo 'worktree-branch-guard: cannot inspect command because python3 is unavailable' >&2
  echo '{}'; exit 0
fi
matches=$(python3 - "$cmd" <<'PY'
import shlex, sys
source = sys.argv[1]
try:
    lexer = shlex.shlex(source, posix=True, punctuation_chars='|;&\n')
    lexer.whitespace = ' \t\r'
    lexer.whitespace_split = True
    words = list(lexer)
except ValueError:
    raise SystemExit(0)
segments, segment = [], []
for word in words:
    if word and set(word) <= set('|;&\n'):
        if segment: segments.append(segment)
        segment = []
    else: segment.append(word)
if segment: segments.append(segment)
for seg in segments:
    for start, word in enumerate(seg):
        if word.rsplit('/', 1)[-1] != 'git': continue
        i, repo = start + 1, ''
        while i < len(seg) and (seg[i] in ('-c', '--git-dir', '--work-tree', '-C') or seg[i].startswith('--git-dir=') or seg[i].startswith('--work-tree=')):
            opt = seg[i]
            if opt in ('-c', '--git-dir', '--work-tree', '-C') and i + 1 < len(seg):
                val = seg[i + 1]; i += 2
                if opt in ('-C', '--work-tree'): repo = val
                elif opt == '--git-dir' and val.endswith('/.git'): repo = val[:-5]
            elif opt.startswith('--git-dir=') or opt.startswith('--work-tree='):
                if opt.startswith('--work-tree='): repo = opt.split('=', 1)[1]
                elif not repo and opt.endswith('/.git'): repo = opt[:-5]
                i += 1
            else: i += 1
        if i >= len(seg): continue
        sub, args = seg[i], seg[i+1:]
        branch = ''
        if sub == 'checkout':
            for j, arg in enumerate(args):
                if arg == '-B' and j+1 < len(args): branch = args[j+1]; break
                if arg.startswith('-B') and len(arg) > 2: branch = arg[2:]; break
        elif sub == 'switch':
            for j, arg in enumerate(args):
                if arg == '-C' and j+1 < len(args): branch = args[j+1]; break
                if arg.startswith('-C') and len(arg) > 2: branch = arg[2:]; break
        elif sub == 'branch':
            for j, arg in enumerate(args):
                if arg in ('-f', '--force') and j+1 < len(args): branch = args[j+1]; break
                if arg.startswith('-f') and len(arg) > 2: branch = arg[2:]; break
        elif sub == 'update-ref' and args and args[0].startswith('refs/heads/'):
            branch = args[0][11:]
        elif sub == 'fetch':
            for arg in args:
                if ':' in arg and not arg.startswith('-'):
                    branch = arg.rsplit(':', 1)[1].removeprefix('+').removeprefix('refs/heads/'); break
        if branch: print(repo + '|' + branch)
PY
)
[ -n "$matches" ] || { echo '{}'; exit 0; }
current=$(pwd -P)
while IFS='|' read -r repo branch; do
  [ -n "$branch" ] || continue
  if [ -n "$repo" ]; then
    case "$repo" in /*) repo_path=$repo ;; *) repo_path="$current/$repo" ;; esac
    repo_path=$(realpath -m "$repo_path")
    root=$(git -C "$repo_path" rev-parse --show-toplevel 2>/dev/null) || continue
  else
    root=$(git rev-parse --show-toplevel 2>/dev/null) || continue
  fi
  root=$(realpath "$root")
  while IFS= read -r line; do
    case "$line" in
      'worktree '*) path=${line#worktree } ;;
      "branch refs/heads/$branch")
        path=$(realpath "$path" 2>/dev/null || printf '%s' "$path")
        if [ "$path" != "$root" ]; then
          jq -nc --arg path "$path" --arg branch "$branch" '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:("Blocked: branch `"+$branch+"` is checked out in another worktree: "+$path+". Perform the operation in that worktree, or detach it with `git checkout --detach` first.")}}'
          exit 0
        fi
        ;;
    esac
  done < <(git -C "$root" worktree list --porcelain 2>/dev/null)
done <<< "$matches"
echo '{}'
