#!/usr/bin/env bash
set -euo pipefail
wt=${1:?usage: identgone.sh <worktree>}
python3 - "$wt" "${ALLOW_IDENT_GONE:-}" <<'PY'
import os, re, subprocess, sys
from collections import Counter

wt, allow_raw = sys.argv[1:]
allow = set(allow_raw.split())
keywords = set('''
const return func import export string false true null undefined let var if else
for while do switch case break continue default class interface type enum public
private protected static readonly async await new this super void int bool byte
package go defer range select chan map struct fallthrough goto make len cap append
nil true false error err function throw try catch finally delete in of instanceof
typeof undefined NaN Infinity extends implements yield from as is any unknown never
number boolean object symbol bigint declare namespace module require
'''.split())

def run(*args):
    return subprocess.check_output(['git', '-C', wt, *args], stderr=subprocess.DEVNULL)

def identifiers(source):
    # Replace comments and quoted strings with spaces, retaining code around them.
    out=[]; i=0; n=len(source); state='code'
    while i<n:
        c=source[i]; nxt=source[i+1] if i+1<n else ''
        if state=='code':
            if c=='/' and nxt=='/': state='line'; out.extend('  '); i+=2; continue
            if c=='/' and nxt=='*': state='block'; out.extend('  '); i+=2; continue
            if c in "'\"`": state={'\'':'single','"':'double','`':'template'}[c]; out.append(' '); i+=1; continue
            out.append(c); i+=1; continue
        if state=='line':
            if c=='\n': state='code'; out.append('\n')
            else: out.append(' ')
            i+=1; continue
        if state=='block':
            if c=='*' and nxt=='/': out.extend('  '); i+=2; state='code'
            else: out.append('\n' if c=='\n' else ' '); i+=1
            continue
        # Single/double quoted content and template literal text are ignored.
        if c=='\\': out.extend('  '); i+=2; continue
        closing = {'single':"'", 'double':'"', 'template':'`'}[state]
        if c==closing: out.append(' '); i+=1; state='code'; continue
        out.append('\n' if c=='\n' else ' '); i+=1
    return Counter(x for x in re.findall(r'(?<![A-Za-z0-9_])[A-Za-z_][A-Za-z0-9_]*', ''.join(out)) if len(x)>=4 and x not in keywords)

def edit_distance_at_most_two(left, right):
    left, right = left.lower(), right.lower()
    if abs(len(left) - len(right)) > 2:
        return False
    previous = list(range(len(right) + 1))
    for i, a in enumerate(left, 1):
        current = [i]
        for j, b in enumerate(right, 1):
            current.append(min(current[-1] + 1, previous[j] + 1, previous[j - 1] + (a != b)))
        if min(current) > 2:
            return False
        previous = current
    return previous[-1] <= 2

paths=run('diff','--name-only','-z','HEAD','--','*.ts','*.tsx','*.go').decode().split('\0')
for path in paths:
    if not path: continue
    try:
        old=run('show',f'HEAD:{path}').decode(errors='replace')
        with open(os.path.join(wt,path), encoding='utf-8', errors='replace') as f: new=f.read()
    except (subprocess.CalledProcessError, OSError):
        continue
    old_names, new_names = identifiers(old), identifiers(new)
    for name in sorted(old_names.keys()-new_names.keys()-allow):
        candidates = [candidate for candidate in sorted(new_names.keys()-old_names.keys()) if (
            (min(len(name), len(candidate)) >= 6 and (name in candidate or candidate in name)) or
            (len(name) >= 6 and len(candidate) >= 6 and edit_distance_at_most_two(name, candidate))
        ) and old_names[name] == new_names[candidate]]
        if candidates:
            print(f'WARN IDENT_GONE {path} {name} -> {candidates[0]}')
        else:
            print(f'NOTE IDENT_GONE {path} {name}')
PY
