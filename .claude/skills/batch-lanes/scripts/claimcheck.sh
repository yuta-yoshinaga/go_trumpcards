#!/usr/bin/env bash
set -euo pipefail
if (($# != 2)); then echo "usage: claimcheck.sh <worktree> <text>" >&2; exit 2; fi
worktree=$1 text=$2
if [[ "$text" == *"claim-check: skip"* ]]; then exit 0; fi
python3 - "$worktree" "$text" <<'PY'
import os, re, subprocess, sys
root, text = sys.argv[1:]
bad = []
def report(kind, token): bad.append(f"UNVERIFIED {kind} {token}")
def grep(token):
    return subprocess.run(["git", "grep", "--untracked", "-F", "-q", "--", token], cwd=root,
                          stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL).returncode == 0
ids = set()
for token in re.findall(r"`([^`]+)`", text):
    if len(token) >= 4 and re.fullmatch(r"[A-Za-z_][A-Za-z0-9_.]*", token) and re.search(r"[A-Z_.]", token):
        ids.add(token.rsplit(".", 1)[-1])
for token in sorted(ids):
    if not grep(token): report("identifier", token)
files = subprocess.run(["git", "ls-files", "-co", "--exclude-standard", "-z"], cwd=root,
                       stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, check=True).stdout
tracked_files = [p.decode("utf-8", "replace") for p in files.split(b"\0") if p]
for path, line in re.findall(r"(?<![\w/])([A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*\.[A-Za-z0-9]+):(\d+)\b", text):
    candidates = [path] if "/" in path else [p for p in tracked_files if p.endswith("/" + path) or p == path]
    found = False
    for candidate in candidates:
        try:
            with open(os.path.join(root, candidate), encoding="utf-8", errors="replace") as f:
                if sum(1 for _ in f) >= int(line):
                    found = True
                    break
        except OSError: pass
    if not found: report("file:line", f"{path}:{line}")
names = set()
for quote, name in re.findall(r"\bit\s*\(\s*(['\"])(.*?)\1", text): names.add(name)
for quote, name in re.findall(r"\bt\.Run\s*\(\s*(['\"])(.*?)\1", text): names.add(name)
names.update(re.findall(r"\bfunc\s+(Test[A-Za-z0-9_]*)\s*\(", text))
if names:
    contents = []
    for path in tracked_files:
        filename = os.path.basename(path)
        if filename.endswith("_test.go") or re.search(r"\.test\.tsx?$", filename):
            try:
                with open(os.path.join(root, path), encoding="utf-8", errors="replace") as f: contents.append(f.read())
            except OSError: pass
    joined = "\n".join(contents)
    for name in sorted(names):
        if name not in joined: report("test", name)
if bad:
    print("\n".join(bad)); sys.exit(1)
PY
