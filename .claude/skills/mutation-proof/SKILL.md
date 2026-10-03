---
name: mutation-proof
description: Prove a test detects a targeted code change by running it once with a literal mutation and again after restoration. Use after writing or receiving a test for a branch, or before claiming mutation verification on a PR.
# Model-invocable: performs a temporary, automatically restored source mutation.
allowed-tools: Read, Grep, Glob, Bash
argument-hint: "<file> <old-literal> <new-literal> -- <test command...>"
---

# Mutation Proof

Use this after adding a test that claims to cover a branch, and before saying a PR was verified
by mutation. The script changes exactly one literal, runs the test, restores the original file,
then runs the test again:

```bash
bash .claude/skills/mutation-proof/scripts/mutate.sh frontend/src/utils/example.ts 'a > 0' 'a < 0' -- bunx vitest run frontend/src/utils/example.test.ts
bash .claude/skills/mutation-proof/scripts/mutate.sh internal/example.go 'return true' 'return false' -- go test ./internal/example
```

Choose a mutation that breaks the behavior asserted by the test: negate a condition (`n > 0` to
`n <= 0`), drop a guard's effect (`if (busy) return` to `if (false) return`), or change a
dependent constant (`limit = 3` to `limit = 4`).

The literal must match exactly once; zero or multiple matches abort because they cannot establish
which behavior was changed. Never remove a null guard if that makes later code throw: use a
behavior-preserving but wrong mutation, such as setting a ref back to its old initial value.
Regex assertions can accidentally accept shifted values: `/3秒/` also matches `63秒`, so add a
boundary to the assertion before using that as proof. The script always restores from its backup;
never undo a mutation with git because the file may contain uncommitted work.

`PROVEN` means the mutated run failed and the restored run passed. `NOT PROVEN` means the test
still passed under mutation. `BROKEN BASELINE` means the restored test also failed. More than the
default three failing tests in Vitest's summary or Go's `--- FAIL` lines triggers a collateral
failure warning: the mutation may have crashed a module instead of testing the intended behavior.
Set `MUTATION_MAX_FAILURES` to change that threshold.
