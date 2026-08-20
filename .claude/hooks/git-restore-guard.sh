#!/usr/bin/env bash
# PreToolUse(Bash) gate: refuse a `git checkout -- <path>` / `git restore <path>`
# that would silently discard uncommitted work.
#
# The sibling rule in settings.json already blocks `git checkout <ref> -- .`,
# because that form once wiped eleven files. It does not catch the single-path
# form, and that form has destroyed work twice in one session:
#
#   1. Six mermaid label fixes in docs/manual/cui/skat.md, reverted by a
#      `git checkout --` that was only meant to undo a deliberately broken
#      test fixture in the same file.
#   2. A `LetItRide() -> LetItRideAction()` doc fix, reverted the same way
#      while running negative controls against the file it lived in.
#
# Both times the command did exactly what it was asked to and the loss was
# silent -- `git checkout --` prints nothing on success, so the next green test
# run looks like confirmation rather than like the fix having vanished.
#
# This is NOT a blanket block: discarding changes is a legitimate thing to want.
# It fires only when a named path actually has uncommitted modifications, which
# is precisely the case where the command is unrecoverable. `git stash push`
# does the same job and keeps the work reachable.
#
# `git restore --staged <path>` only rewrites the index, so it is left alone.
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

# Split a command fragment into arguments the way a shell would quote them, without
# evaluating anything. Plain word splitting loses `"docs/my file.md"` to two bogus
# tokens, neither of which matches a real path, so `git status` comes back clean for
# both and the guard waves the command through -- failing OPEN, which is the one
# way a guard must never fail. xargs honours quotes and backslashes and performs no
# substitution, so nothing in the command string can execute here.
split_args() {
  printf '%s' "$1" | xargs -n1 2>/dev/null
}

# Collect the paths a restoring command would overwrite. Empty = not our business.
paths=()
parse_failed=0
case "$cmd" in
  # `git checkout ... -- <paths>` (with or without a ref before the --)
  *git\ checkout\ *--\ *)
    after_sep=${cmd#*--\ }
    mapfile -t paths < <(split_args "$after_sep")
    [ -n "$after_sep" ] && [ ${#paths[@]} -eq 0 ] && parse_failed=1
    ;;
  # `git restore <paths>`, unless it only touches the index.
  *git\ restore\ *)
    case "$cmd" in
      *--staged*)
        # --staged alone is index-only and safe; --staged --worktree is not.
        case "$cmd" in *--worktree*) ;; *) echo '{}'; exit 0 ;; esac
        ;;
    esac
    rest=${cmd#*git restore }
    mapfile -t tokens < <(split_args "$rest")
    [ -n "$rest" ] && [ ${#tokens[@]} -eq 0 ] && parse_failed=1
    for tok in "${tokens[@]}"; do
      case "$tok" in -*) continue ;; esac
      paths+=("$tok")
    done
    ;;
  *)
    echo '{}'; exit 0
    ;;
esac

# Unbalanced quotes and the like leave xargs with nothing. Refuse rather than guess:
# this is a data-loss guard, so an unparseable restore is exactly the case where
# waving it through is worst.
if [ "$parse_failed" -eq 1 ]; then
  jq -nc '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: ("Blocked: could not parse the paths out of this restore command (unbalanced quotes?), " +
                 "so it cannot be checked for uncommitted work. Refusing to guess -- rerun with simple " +
                 "quoting, or use `git stash push -- <paths>` which is recoverable either way.")
    }
  }'
  exit 0
fi

[ ${#paths[@]} -eq 0 ] && { echo '{}'; exit 0; }

# A path is at risk when git reports a worktree modification for it. The second
# column of --porcelain is the worktree status; ' ' there means the change is
# staged only, which `git checkout --` would still overwrite, so treat any
# non-clean entry as at risk.
at_risk=""
for p in "${paths[@]}"; do
  status=$(git status --porcelain -- "$p" 2>/dev/null)
  [ -n "$status" ] && at_risk+="${status}"$'\n'
done

if [ -z "$at_risk" ]; then
  echo '{}'
  exit 0
fi

jq -nc --arg files "$(printf '%s' "$at_risk" | sed '/^$/d')" '{
  hookSpecificOutput: {
    hookEventName: "PreToolUse",
    permissionDecision: "deny",
    permissionDecisionReason: ("Blocked: this would discard uncommitted changes with no way back. " +
               "It prints nothing on success, so the loss is silent and the next green " +
               "test run reads like confirmation.\n\nAt risk:\n" + $files +
               "\n\nUse `git stash push -- <paths>` instead (recoverable via `git stash pop`), " +
               "or commit first. If you are restoring a file you deliberately broke for a " +
               "test, stash before breaking it rather than restoring after.")
  }
}'
