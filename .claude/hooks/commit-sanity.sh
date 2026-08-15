#!/usr/bin/env bash
# PreToolUse(Bash) gate: refuse a `git commit` that would record nonsense.
#
# Two failure modes this catches, both of which have reached a branch before:
#
#   1. Conflict markers in staged content. `<<<<<<< HEAD` fails neither tsc nor
#      vitest -- tsc keeps going and vitest never loads the affected module, so
#      both stay green. The only gate that fails is the E2E production build,
#      minutes into CI.
#   2. A commit that stages nothing. `git commit` on an empty index (or after a
#      `git add` that silently matched no paths) produces a commit whose
#      `git show --stat` is zero lines, which then gets pushed and reported as
#      shipped work.
#
# Only `<<<<<<<` and `>>>>>>>` are matched, never `=======`: a row of equals signs
# is a legitimate Markdown heading underline and this repo has many. Both markers
# matched here must be followed by whitespace (git writes `<<<<<<< HEAD`), so a
# line of angle brackets in prose or a shell heredoc does not trip it either.
#
# Reads the hook payload on stdin, writes a hook decision object on stdout.
#
# Decision shape: this guard denies ONE tool call, so it emits
# hookSpecificOutput.permissionDecision = "deny". It must never emit
# {"continue": false}, which is the kill switch: that ends the whole turn, and
# its stopReason is documented as "Not shown to Claude" -- so the agent is
# stopped without ever learning why, and the Stop hook does not run either.
# Every guard in this repo used the kill switch until 2026-08-16; that is what
# was behind the recurring "a hook block ended my turn" failures.
set -uo pipefail

payload=$(cat)
cmd=$(printf '%s' "$payload" | jq -r '.tool_input.command // ""')

case "$cmd" in
  git\ commit*) ;;
  *) echo '{}'; exit 0 ;;
esac

# `--amend` legitimately stages nothing when only the message is being rewritten,
# and `--allow-empty` is an explicit request for the thing this guard blocks.
allows_empty=0
case "$cmd" in
  *--amend*|*--allow-empty*) allows_empty=1 ;;
esac

if [ -z "$(git diff --cached --name-only)" ] && [ "$allows_empty" -eq 0 ]; then
  jq -nc '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: ("Blocked: nothing is staged, so this commit would record zero changes. " +
                 "Check `git status` -- a `git add` that matched no paths is silent. " +
                 "Pass --allow-empty if an empty commit really is what you want.")
    }
  }'
  exit 0
fi

# Walk the staged diff, tracking the current +++ header so a hit can be named.
conflicted=$(git diff --cached -U0 | awk '
  /^\+\+\+ b\// { file = substr($0, 7); next }
  /^\+(<<<<<<<|>>>>>>>)[ \t]/ { if (!(file in seen)) { seen[file] = 1; print file } }
')

if [ -n "$conflicted" ]; then
  jq -nc --arg files "$conflicted" '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: ("Blocked: staged content contains merge conflict markers. Neither tsc nor " +
                 "vitest fails on these -- only the E2E production build does, minutes into CI. " +
                 "Resolve them in:\n" + $files)
    }
  }'
  exit 0
fi

echo '{}'
