---
name: coverage-gate
description: Measure coverage for only the files this branch changed and report which fall under 80%, before pushing. Use when finishing a change, before opening a PR, or when asked whether the coverage standard is met ("カバレッジ足りてる?", "check coverage", "ready to push?").
disable-model-invocation: true
allowed-tools: Read, Grep, Glob, Bash
argument-hint: "[base-ref]"
---

# Coverage Gate

Self-review item 5 — "branch coverage is 80%+ for modified packages" — is the only item on the
checklist with no command behind it. Codecov enforces the same 80% on project and patch, but
only after a push, which makes the pre-push check either manual arithmetic or a full-suite run
nobody waits for.

This runs both suites **scoped to the changed files**, so the answer takes seconds.

```bash
bash .claude/skills/coverage-gate/scripts/coverage-gate.sh          # vs develop
bash .claude/skills/coverage-gate/scripts/coverage-gate.sh master   # vs another base
```

Exit status is 0 when everything in scope clears 80%, 1 otherwise.

## What it measures, precisely

| Suite | Number reported | Scope |
|---|---|---|
| Go | **statement** coverage, `go test -cover` per changed package | `internal/**`, `api/**`, minus `internal/infrastructure/` |
| Frontend | **branch** coverage, vitest v8 provider | `frontend/src/{api,components,pages,utils}` |

Two things worth being exact about:

- **Go has no branch-coverage mode.** The toolchain reports statements, and so does Codecov,
  so the gate here is the same number CI gates on — but it is not C1, and a package at 85%
  statements can still leave a branch untested. When a change is branch-heavy, read the
  uncovered lines (`go test -coverprofile=/tmp/c.out ./pkg && go tool cover -html=/tmp/c.out`)
  rather than trusting the percentage.
- **`cmd/` and `internal/infrastructure/` are excluded** — `codecov.yml` ignores them, so a
  number for them would be one nobody gates on.

Frontend files are measured one at a time: each changed `X.tsx` is measured by running only
`X.test.tsx`. This is deliberate — the full suite with coverage runs past ten minutes locally,
and the question being asked is about the changed file, not the repo.

## Reading the output

```
== Go ==
  internal/usecase                                             76.4%  UNDER 80%
== Frontend (branch coverage) ==
  src/pages/HeartsPage.tsx                                     NO TEST FILE
```

- **UNDER 80%** — add tests before pushing; Codecov's patch target will fail otherwise.
- **NO TEST FILE** — a source file with no `*.test.tsx?` next to it. Untracked files are
  included in the scan, so a brand-new file with no test is caught here rather than in review.
- **NO RESULT** — the suite failed to run. That is a broken test, not a coverage problem;
  fix it and re-run.

## What this does not do

It does not run the full suite, and passing it is not a substitute for CI. Tests and lint
belong on GitHub Actions; this is the narrow pre-push question of whether the *changed* files
carry tests, answered locally because CI answers it too late to act on.
