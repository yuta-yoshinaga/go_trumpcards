#!/usr/bin/env bash
# PreToolUse(Bash): protect branches checked out in another worktree.
set -uo pipefail
payload=$(cat)
cmd=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""')

IFS='|' read -r repo branch < <(python3 - "$cmd" <<'PY'
import shlex, sys
try:
 lexer=shlex.shlex(sys.argv[1], posix=True, punctuation_chars='|;&'); lexer.whitespace_split=True; w=list(lexer)
except ValueError: print('|'); raise SystemExit
repo=''; i=0
while i < len(w):
    if w[i] == 'git':
        i += 1
        while i < len(w) and w[i].startswith('-'):
            if w[i] == '-C' and i+1 < len(w): repo=w[i+1]; i+=2; continue
            i+=1
        if i >= len(w): break
        sub=w[i]; a=w[i+1:]
        branch=''
        if sub == 'checkout' and '-B' in a:
            j=a.index('-B'); branch=a[j+1] if j+1<len(a) else ''
        elif sub == 'switch' and '-C' in a:
            j=a.index('-C'); branch=a[j+1] if j+1<len(a) else ''
        elif sub == 'branch' and ('-f' in a or '--force' in a):
            j=next((j for j,x in enumerate(a) if x in ('-f','--force')),0); branch=next((x for x in a[j+1:] if not x.startswith('-')), '')
        elif sub == 'update-ref' and a and a[0].startswith('refs/heads/'):
            branch=a[0][11:]
        elif sub == 'fetch':
            for x in a:
                if ':' in x and not x.startswith('-'):
                    branch=x.rsplit(':',1)[1].removeprefix('+')
                    if branch.startswith('refs/heads/'): branch=branch[11:]
                    break
        print(repo+'|'+branch); raise SystemExit
    i+=1
print('|')
PY
)
[ -n "$branch" ] || { echo '{}'; exit 0; }
current=$(git ${repo:+-C "$repo"} rev-parse --show-toplevel 2>/dev/null) || { echo '{}'; exit 0; }
while IFS= read -r line; do
  case "$line" in
    'worktree '*) path=${line#worktree } ;;
    "branch refs/heads/$branch")
      if [ "$path" != "$current" ]; then
        jq -nc --arg path "$path" --arg branch "$branch" '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:("Blocked: branch `"+$branch+"` is checked out in another worktree: "+$path+". Perform the operation in that worktree, or detach it with `git checkout --detach` first.")}}'
        exit 0
      fi
      ;;
  esac
done < <(git ${repo:+-C "$repo"} worktree list --porcelain 2>/dev/null)
echo '{}'
