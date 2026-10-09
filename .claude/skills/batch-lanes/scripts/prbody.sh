#!/usr/bin/env bash
# PR body assembly for ship.sh, kept free of side effects so the hook tests can source it.
#
# ship.sh's 4th argument is normally summary lines, which it wraps in "## Summary",
# "Closes #n", "## Test plan" and the Claude Code footer. A caller that writes the
# whole body itself (its own "## " headings) used to get every one of those twice:
# two "## Summary" headings, two "Closes" lines and two footers on each PR (2026-10-09,
# #11594-#11605). A body that already has a "## " heading is used as written; only
# a missing "Closes/Part of #n" and a missing footer are added.

# prbody_is_full <text>: true when the text is a complete body (has its own "## " heading).
prbody_is_full() {
  grep -qE '^## ' <<<"$1"
}

# prbody_has_ref <issue#> <text>: true when the text already closes or references the issue.
prbody_has_ref() {
  grep -qiE "(closes|fixes|resolves|part of) #$1([^0-9]|$)" <<<"$2"
}

# ship_pr_body <issue#> <summary-or-body> <frontend test-plan line> <go test-plan line>
# Prints the PR body. Reads BATCH_DELEGATED_TO and CLAUDE_SESSION_URL like ship.sh.
ship_pr_body() {
  local n=$1 summ=$2 tp_fe=$3 tp_go=$4 footer=''
  if [[ -n "${CLAUDE_SESSION_URL:-}" ]]; then
    footer="🤖 Generated with [Claude Code](https://claude.com/claude-code)

$CLAUDE_SESSION_URL"
  fi
  if prbody_is_full "$summ"; then
    prbody_has_ref "$n" "$summ" || printf 'Closes #%s\n\n' "$n"
    printf '%s\n' "$summ"
    if [[ -n "$footer" ]] && ! grep -qF 'Generated with [Claude Code]' <<<"$summ"; then
      printf '\n%s\n' "$footer"
    fi
    return
  fi
  local closes="Closes #$n"
  prbody_has_ref "$n" "$summ" && closes=''
  cat <<M
## Summary
$summ

$closes

## Test plan
$tp_fe
- [x] \`bun run check\` / \`bun run typecheck\`
$tp_go
- [ ] CI green

${BATCH_DELEGATED_TO:+Delegated-To: $BATCH_DELEGATED_TO}

$footer
M
}
