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
changed diff lines. A printed row is `<file>:<line>: <branch type> [counts] | <source>`. Any
zero in `[counts]` means that branch outcome was not taken; the script exits 1 if it finds one.
`NO TEST FILE` means the source has no sibling `.test.ts` or `.test.tsx` file.

Remedy gaps in this order:

1. Delete an unreachable defence. Common examples are `return state?.value ?? 0` when the caller
   guarantees `state`, `state?.value` after `if (!state) return`, and a nested ternary arm that
   cannot occur given the surrounding condition.
2. Move a helper below `if (!state) return` so optional chaining is unnecessary.
3. Add a test that exercises the reachable branch.

Go coverage is statement coverage and has no partial-line branch metric, so this skill is
frontend-only. For nested ternaries, `the [0, n] entry may be the INNER conditional`: the second
`cond-expr` of a nested ternary can be reported on the same line.
