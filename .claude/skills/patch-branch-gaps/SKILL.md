---
name: patch-branch-gaps
description: Find uncovered frontend branches on changed lines before pushing, including partial lines Codecov counts as misses. Use when finishing a frontend change, before pushing, or when codecov/patch reports a miss.
# Model-invocable: scoped measurement of changed frontend branches.
allowed-tools: Read, Grep, Glob, Bash
argument-hint: "[base-ref]"
---

# Patch Branch Gaps

Run this after a frontend change, before pushing, or when Codecov's patch check fails:

```bash
bash .claude/skills/patch-branch-gaps/scripts/patch-branch-gaps.sh
bash .claude/skills/patch-branch-gaps/scripts/patch-branch-gaps.sh origin/main
```

It runs only each changed source file's sibling test and inspects branches located on added or
changed diff lines. A new uncovered branch is printed as
`<file>:<line>: <branch type> [counts] | <source>`; a branch whose trimmed source text is already
a prefix of a trimmed line in `<base>:frontend/<file>` (the base defaults to the merge-base with `origin/develop`) (at least 8 characters) is printed
as `PREEXISTING <file>:<line>: <branch type> [counts] | <source>`. The summary reports new and
pre-existing counts separately. The script exits 1 for new uncovered branches or measurement
errors such as `NO RESULT`, and exits 0 when it finds only pre-existing branches. `NO TEST FILE`
means the source has no sibling `.test.ts` or `.test.tsx` file.

Remedy gaps in this order:

1. Delete an unreachable defence. Common examples are `return state?.value ?? 0` when the caller
   guarantees `state`, `state?.value` after `if (!state) return`, and a nested ternary arm that
   cannot occur given the surrounding condition.
2. Move a helper below `if (!state) return` so optional chaining is unnecessary.
3. Add a test that exercises the reachable branch.

Go coverage is statement coverage and has no partial-line branch metric, so this skill is
frontend-only. For nested ternaries, `the [0, n] entry may be the INNER conditional`: the second
`cond-expr` of a nested ternary can be reported on the same line.
