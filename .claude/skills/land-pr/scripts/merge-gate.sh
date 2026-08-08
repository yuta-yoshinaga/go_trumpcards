#!/usr/bin/env bash
# Report whether a PR is safe to merge. Exits 0 only when every condition holds.
#
# Usage: merge-gate.sh <pr-number> [expected-head-sha]
#
# Prints a line per condition and a verdict. Never merges anything -- deciding and
# doing are separated on purpose, so the decision can be read before it is acted on.
set -uo pipefail

pr=${1:?usage: merge-gate.sh <pr-number> [expected-head-sha]}
expected=${2:-$(git rev-parse HEAD 2>/dev/null)}

fail=0
say() { printf '  %-6s %s\n' "$1" "$2"; }

echo "merge gate: PR #$pr"

# 1. The PR head must be the commit whose checks we are about to read. Without this,
#    a `gh pr checks` run moments after a push reports the PREVIOUS commit's green
#    result -- the new run has not registered yet. That exact hole merged two PRs
#    early (#4345, #4347).
head=$(gh pr view "$pr" --json headRefOid --jq .headRefOid 2>/dev/null)
if [ -z "$head" ]; then
  say FAIL "cannot read PR #$pr (gh auth? wrong number?)"
  exit 1
fi
if [ "$head" = "$expected" ]; then
  say ok "head matches the commit under test ($head)"
else
  say FAIL "PR head $head != expected $expected -- a newer push exists, or you are on the wrong commit"
  fail=1
fi

# 2. Zero pending. Checked separately from "no failures": a stale-commit read can
#    show zero pending while the real run has not started.
pending=$(gh pr checks "$pr" --json bucket --jq '[.[]|select(.bucket=="pending")]|length' 2>/dev/null || echo unknown)
total=$(gh pr checks "$pr" --json bucket --jq 'length' 2>/dev/null || echo 0)
if [ "$pending" = "0" ]; then
  say ok "no pending checks ($total total)"
else
  say FAIL "$pending check(s) still pending"
  fail=1
fi

# 3. Zero failures.
failing=$(gh pr checks "$pr" --json bucket,name --jq '[.[]|select(.bucket=="fail")|.name]|join(", ")' 2>/dev/null)
if [ -z "$failing" ]; then
  say ok "no failing checks"
else
  say FAIL "failing: $failing"
  fail=1
fi

# 4. Sanity floor on the number of checks. An empty array satisfies every "all
#    green" test ever written, so assert the run actually reported something.
#    Deliberately low: the count varies (mention-review does not always register).
if [ "$total" -ge 10 ] 2>/dev/null; then
  say ok "check count above floor ($total >= 10)"
else
  say FAIL "only $total checks reported -- the run has not registered yet, or gh returned nothing"
  fail=1
fi

# 5. Review comments must be READ, not counted. Auto-review does not re-run on
#    push, so a green tick after a fix-up commit does not mean the fix-up was
#    reviewed. This cannot be automated; surface it and make the human look.
reviews=$(gh pr view "$pr" --json comments --jq '[.comments[]|select(.author.login=="github-actions")]|length' 2>/dev/null || echo 0)
say note "$reviews bot comment(s) -- READ them; auto-review does NOT re-run on push,"
say note "  so a fix-up pushed after the review has not been reviewed at all"
gh api "repos/{owner}/{repo}/pulls/$pr/comments" --jq '.[]|"  inline: \(.path):\(.line // "outdated") \(.user.login)"' 2>/dev/null | head -20

if [ "$fail" -ne 0 ]; then
  echo "VERDICT: BLOCKED"
  exit 1
fi
echo "VERDICT: gate passed -- merge only after reading the reviews above"
