---
name: improve-issue
version: 1.0.0
description: Implement ONE improvement/UX GitHub issue end-to-end — branch, TDD, PR (Closes #N), watch CI, read the FULL auto-review, address every finding, squash-merge, sync. Use for "issueに着手して", "#NNNN を対応して", "issueをやってマージまで", per-issue improvement work.
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Grep
  - Glob
  - AskUserQuestion
# Model-invocable: same reasoning as new-game — improve-batch exists to run this per
# issue, so gating it makes the orchestrator inoperable by design. The entry point
# (improve-batch) stays explicit-invocation only.
---

# improve-issue — one issue, branch → merge

Codifies the loop used to clear the **#2238–#2292** per-game improvement batch
(20 PRs #2376–#2395, all merged green). Reuse it to take a single
improvement/UX/a11y issue from open → merged with reviews addressed.

**Input:** an issue number (e.g. `/improve-issue 2248`). If none is given, pick
the lowest-effort still-open issue from the batch the user names.

> This is the *driver*. The repo rules in [`CLAUDE.md`](../../../CLAUDE.md),
> [`frontend/CLAUDE.md`](../../../frontend/CLAUDE.md) and the user's git/PR
> feedback memories are authoritative — this skill encodes the *sequence* and
> the *gotchas* that aren't obvious from those docs alone.

## The loop (do these in order)

1. **Read the issue & reality-check it.**
   `gh issue view <#> --json title,body`. Then **verify the premise against the
   actual code** before writing anything. The propose-batch occasionally
   misreads code → the work is already done or impossible. If so, **do not
   fabricate** — comment with evidence (cite files/lines) and
   `gh issue close <#> --reason "not planned"`. (Precedents: #2269, #2245, #2272.)

2. **Branch from fresh develop — BEFORE editing.**
   ```sh
   git checkout develop && git pull origin develop
   git checkout -b feat/<#>-<slug>   # ALWAYS branch before the first edit
   ```
   Never edit on `develop`. Never `git add -A` / `git add .` (the worktree has
   local-only files: untracked skills, node_modules cache, build binaries) —
   `git add` only the exact files you changed.

3. **Implement with TDD** (Red → Green → Refactor). Match the surrounding code.
   See **Patterns** below. Always ship the test in the same commit.

4. **Local gates** (frontend):
   ```sh
   cd frontend && bun run test -- --run src/pages/<X>Page.test.tsx
   cd frontend && bun run check          # biome + design-tokens (NOT tsc — see gotchas)
   ```
   For Go changes also: `goimports -w <file>` (install with
   `go install golang.org/x/tools/cmd/goimports@latest`; ensure `$(go env GOPATH)/bin`
   is on `PATH`), `go vet -tags test ./internal/adapter/presenter/` and run the focused package
   test (`internal/domain` OOMs locally — CI-gate it; `adapter/presenter` is fine).

5. **Commit (Conventional Commits), push, open the PR.**
   ```sh
   git add <exact files>; git commit -m "feat(<game>): <imperative>\n\n<body>"
   git push -u origin feat/<#>-<slug>
   gh pr create --base develop --title "..." --body "...Closes #<#>"
   ```
   PR body **must** close the issue (`Closes #<#>`) and include a test plan.

6. **Watch CI** (`gh pr checks <pr> --watch --interval 30`, run in background).
   Triage failures — see **CI triage** below.

7. **Read the FULL auto-review before merging — not just the verdict.**
   `gh pr view <pr> --json comments` + inline:
   `gh api repos/<owner>/<repo>/pulls/<pr>/comments`. Cover **every** section
   (all severities + Gemini). Fix every 🔴 Must-Fix and 🟡 Should-Fix, plus any
   codecov gap; apply cheap nits too. Push fixups (they re-run CI).
   Note: auto-review fires on PR **open only** (not on later pushes), so the
   first review is the one to satisfy.

8. **Merge when fully green, then sync.**
   ```sh
   gh pr merge <pr> --squash --delete-branch
   git checkout develop && git pull origin develop && git branch -D feat/<#>-<slug>
   ```

9. **Record progress** in the running memory tally
   (`memory/project_issues_*.md` + `MEMORY.md`) so the next run has context.

## CI triage (read the log — don't guess)

- **Fast E2E failure (<~1 min)** = a **tsc build error**, not a flake (the E2E
  job builds first). `gh run view --job <id> --log-failed`. Common tsc-only
  errors `bun run check` misses: optional-index (`obj[maybeUndef]`), a
  `HintResult` test mock missing `targetAction`, missing union members.
- **Long E2E failure (~15–20 min) with `Target page, context or browser has
  been closed`** = the recurring **runner-crash flake** (e.g. poker.spec /
  paigow). `231 passed, 1 failed` confirms it. `gh run rerun <run> --failed`;
  if it recurs, full `gh run rerun <run>` for a fresh runner. develop's E2E is
  green, so it is never your PR.
- **golangci-lint binary download HTTP 504** = infra; rerun.
- **codecov/patch < 80%** = a new branch is untested. codecov counts **each
  ternary/`&&` branch and each `?.`/`??`** separately → add a test per branch
  (the eligible **and** the ineligible path, empty-array edges, clamp limits).
  Prefer extracting a pure util (testable in isolation) and removing dead
  defensive guards over leaving partial branches.

## Patterns that work (pick by issue)

- **Single-page frontend UX**: edit `src/pages/<X>Page.tsx` + its `.test.tsx` +
  `i18n/locales/{ja,en}/<x>.json` (keep **key-for-key parity** — the parity
  checker runs in CI). TDD the page test.
- **"validate selection / derive value"**: extract a pure util in
  `src/utils/<x>.ts` mirroring the Go domain logic (verify field names against
  the Go source), with its own `.test.ts`. Keeps codecov happy and logic
  verifiable. (e.g. `tonkMeldIndices`, `osmosisRules`, `fivehundredBidValue`.)
- **Shared hand/board highlight**: `PlayerHandSection`/`MobileHandGrid` take an
  optional `highlightIndices?` prop (backward-compatible). Card buttons set
  `boxShadow` **inline**, so a Tailwind `ring-*` is overridden → highlight via
  an inline style helper (`highlightCardStyle`/`playableCardStyle`), not a ring.
- **Transient feedback / live value**: `useState` + `useRef` timer cleared in a
  `useEffect(() => () => clearTimeout(...), [])`; track previous value in a ref
  to fire only on a real transition. Add `aria-live="polite"` to live readouts;
  clamp `aria-valuenow` into `[min,max]`. Test timers with `vi.useFakeTimers()`
  + `vi.waitFor` (NOT RTL `waitFor`) and wrap `vi.advanceTimersByTime` in `act`.
- **Hooks before early return**: declare every hook (incl. `useGameHint`,
  `useSolitaireDragDrop`, refs/effects) **above** any `if (!state) return` —
  biome `useHookAtTopLevel` fails otherwise.
- **CUI / Go presenter**: mirror an existing `HintOutput` (Truco/Briscola); add
  i18n keys to BOTH `internal/i18n/locales/{ja,en}/<game>.json`; assert output
  with `assert.Contains`.

## Conventions baked in

- Design tokens only (`text-ds-*`, `bg-ds-*`); raw palette and `text-white/N`
  opacity variants are rejected by `check-design-tokens.mjs`.
- biome requires `${n.toString()}` in template literals (numbers); `clearTimeout`
  needs `?? undefined` (rejects `null`) — both are real tsc/biome constraints
  reviewers may wrongly call redundant.
- Drop comments that merely echo a constant name; remove dead branches; align
  i18n `defaultValue` strings with the locale files.
