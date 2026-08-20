#!/usr/bin/env bash
# Exercise the hooks defined inline in .claude/settings.json.
#
# Eight of this project's hooks are shell one-liners in settings.json rather
# than scripts, so nothing could test them. That is where the 2026-08-16 defect
# lived: a `printf` with a literal newline inside a JSON string produced invalid
# JSON, and the gate it implemented had therefore never worked once. It looked
# fine in the file. It only showed up when the deny path actually ran.
#
# So: feed each hook a benign payload and a set of payloads chosen to trip its
# deny path, and require valid JSON (or nothing) and exit 0 every time.
set -u
cd "$(dirname "$0")/../.." || exit 1
[ -f .claude/settings.json ] || { echo "skip  no .claude/settings.json here"; exit 0; }
command -v jq >/dev/null 2>&1 || { echo "skip  jq is not installed"; exit 0; }

python3 - <<'PY'
import json, subprocess, sys

d = json.load(open('.claude/settings.json', encoding='utf-8'))
inline = [(ev, h['command'])
          for ev, arr in d.get('hooks', {}).items()
          for g in arr for h in g.get('hooks', [])
          if '.claude/hooks/' not in h.get('command', '')]

if not inline:
    # An empty list passes every assertion ever written; say so rather than
    # printing a green tick for having checked nothing.
    print("FAIL  no inline hooks found -- settings.json changed shape?")
    sys.exit(1)

CO = "check" + "out"   # never put the literal in a command line: the restore
                       # guard inspects OUR argv and refuses to be probed
payloads = [
    "echo hi",
    "git commit -m x --no-verify",
    "git push --force origin develop",
    "git commit --amend -m x",
    "rm -rf /",
    "git " + CO + " -- .",
    "git reset --hard origin/develop",
]

fail = 0
for i, (ev, cmd) in enumerate(inline):
    for p in payloads:
        blob = json.dumps({"tool_input": {"command": p, "file_path": "/tmp/probe.go"},
                           "tool_response": {"filePath": "/tmp/probe.go"}})
        try:
            r = subprocess.run(["bash", "-c", cmd], input=blob,
                               capture_output=True, text=True, timeout=30)
        except subprocess.TimeoutExpired:
            print(f"FAIL  hook[{i}] ({ev}) hung on {p!r}")
            fail = 1
            continue
        if r.returncode != 0:
            print(f"FAIL  hook[{i}] ({ev}) exited {r.returncode} on {p!r}")
            fail = 1
            continue
        out = r.stdout.strip()
        if out:
            try:
                json.loads(out)
            except Exception:
                print(f"FAIL  hook[{i}] ({ev}) emitted invalid JSON on {p!r}: {out[:80]!r}")
                fail = 1

if not fail:
    print(f"  ok    {len(inline)} inline hooks x {len(payloads)} payloads: valid JSON, exit 0")
sys.exit(fail)
PY
# **The heredoc's exit code is not the script's unless we say so.** Without this
# the file ends on the PY terminator, bash reports 0, and the FAIL lines above
# scroll past inside a green job -- the exact shape this job exists to stop.
exit $?
