---
name: new-game
description: Scaffold and wire a new card game across every backend, frontend, worker, doc, and count-assertion touchpoint. Use when the user wants to add a new game (e.g. "add a new game", "implement <game>", "/new-game <name> <category>"). Drives the full New Game Addition Checklist so no registration point or hardcoded count is missed.
# Model-invocable: this is a STEP, not an entry point. The human starts the batch or
# the standing new-game loop; blocking the step there does not prevent the commits and
# PRs — it only removes the checklist that keeps them correct, and leaves the
# orchestrator (improve-batch / the autonomous loop) unable to do the one thing it is for.
---

# New Game

Add a new card game end to end without the recurring first-push failures (forgotten count
assertions, wrong worker bucket, missing registration touchpoint). This skill is the
*driver*; the authoritative step list lives in
[`docs/new-game-checklist.md`](../../../docs/new-game-checklist.md) — read it first and treat
it as the source of truth. This file adds the moving parts that bite most often.

## Inputs

- `<name>` — lowercase game id, no spaces (e.g. `fivehundred`, `sixcardgolf`). Used as the
  registry key, API route, i18n filename, and worker route. Must be unique.
- `<category>` — exactly one of `casino` | `classic` | `solo`. **This is a binary-size
  bucket, not a user-facing taxonomy.** It pins the game to one Cloudflare Worker WASM
  binary. Pick the category whose worker still has gzip headroom under the 1 MB free-tier
  limit, *not* the one that "feels" right thematically.
  - ⚠️ The `classic` worker is historically at/near the 1 MB limit. Recent games
    (Scopa, Barbu, Macau, Tien Len, Wasp, Thirty-One, Osmosis, Five Hundred) were all
    bucketed `solo` or `casino` for headroom even when thematically trick-taking. When in
    doubt, prefer `solo` (most headroom) and confirm with `make build-worker-<category>`.
- `<issue#>` (optional) — GitHub issue this closes; include `Closes #<n>` in the PR body.

If `<name>` or `<category>` is missing, ask before scaffolding.

## Workflow

1. **Research & reuse first.** Before writing a domain, check whether an existing game is a
   near-clone. The repo has many derived games (Wasp = Scorpion + empty-column rule;
   Big O = Omaha with `holeCards=5`; Tien Len = Big Two variant). Cloning an existing
   domain + adjusting one rule is the dominant successful pattern — search
   `internal/domain/` for the closest analog and read it before starting net-new code.

2. **TDD per layer (Red → Green → Refactor).** Tests ship in the same commit. Domain →
   usecase → adapter → frontend, each with its test file. Target **80%+ branch coverage**
   on every new package (`cmd/` and `internal/infrastructure/` are exempt).

3. **Walk the checklist.** Open [`docs/new-game-checklist.md`](../../../docs/new-game-checklist.md)
   and complete every item (backend, frontend, docs, verification). Do not skip the
   self-audit block at the bottom.

4. **Bump ALL count assertions** (the #1 first-push failure). Adding one game changes these
   hardcoded totals. Run the bundled audit to see, in one shot, the source-of-truth count
   (`RegisterKVGame` calls per category) against every assertion and exactly which are stale:

   ```sh
   bash .claude/skills/new-game/scripts/count-audit.sh
   ```

   It is read-only (edits nothing) and exits non-zero on any mismatch. Bump every ❌ line to
   the source-of-truth value, then re-run until it prints all ✅. The assertions it covers:

   | File | What to bump |
   |------|--------------|
   | `internal/infrastructure/games/registry_test.go` | the matching `expectedCasino` / `expectedClassic` / `expectedSolo` const (line 16–18). `expectedTotal` is derived — do not touch it. |
   | `frontend/src/hooks/useTutorialProgress.test.ts` | `totalCount).toBe(N)` (~line 12) |
   | `frontend/src/components/tutorial/TutorialProgressPanel.test.tsx` | **three** assertions: `getByText(/N/)`, `links.length === N`, `incompleteMarkers.length === N` (~lines 22, 36, 49) |

   The three frontend assertions are **tsc-only** — `bun run check` (biome) does not catch a
   stale count, so without this audit they slip through to a failed CI run. The Go
   `expected<Category>` const must match the **count of `RegisterKVGame` calls in that
   category's sub-package**, and `TestWorkerRegistrationsCoverAllGames` parses the sources to
   enforce registry ↔ worker agreement (ADR-0031). A mismatch fails the build.

5. **Registration ordering gotcha.** `TestRegistryMatchesCLI` requires the order of entries
   in `internal/infrastructure/ui/GameManager.go` `gameRegistry` to match the order in
   `internal/infrastructure/games/registry.go`. Append the new game **last in both**.

6. **The four backend registration points must all agree on `<name>` and `<category>`:**
   - `registry.go` → `{Name, Category, Description}`
   - `games_server.go` → `BindWebControllerFor("<name>", …)`
   - `internal/infrastructure/games/<category>/<category>.go` → `RegisterKVGame("<name>", games.Category…, …)`
   - `frontend/src/api/gameApi.ts` `workerUrl` → `"<name>": "<category>"`

7. **Frontend route requires a `profile`.** Every `gameRoutes.ts` entry needs the
   `profile: GameProfile` field (4 axes; see `discoverAxes.ts` for option order) or tsc
   rejects the file. Also add `discover.blurb.<page-kebab>` + `discover.stretch_blurb.<page-kebab>`
   in both `i18n/locales/{ja,en}/discover.json`.

8. **Verify, respecting this box's memory limits.** Per the project's RAM constraints,
   **never run heavy builds/tests in parallel** — run one at a time and let CI gate the
   ones that OOM locally (`internal/domain` test pkg, full `golangci-lint`, `bun run build`
   tsc, E2E, tinygo worker-size). What *does* run locally: targeted `go test` on the new
   non-domain packages, `goimports -w`, `biome check`, `vitest`, host `go build ./...`,
   and `make build-worker-<category>` for the size check.

9. **Gate before committing.** First re-run `bash .claude/skills/new-game/scripts/count-audit.sh`
   — it must print all ✅. Then invoke the `game-registration-checker` subagent (Agent tool,
   `subagent_type: "game-registration-checker"`) with the new `<name>` and `<category>`. It
   greps every touchpoint and diffs the count assertions read-only — catching the exact
   failures above *before* the expensive, OOM-prone CI round-trip. Fix anything either flags,
   then commit.

10. **Docs in the same commit.** README.md, CLAUDE.md (games list), docs/games.md,
    docs/architecture.md (+ endpoint count), api/openapi.yaml, and both
    `docs/manual/{cui,web}/<game>.md` (Mermaid flowchart mandatory; Mermaid state-diagram
    notes must use `=` not `:`). Register the manuals in `manualTexts.ts` (Web) and
    `cuiManualTexts.ts` (CUI).

## Commit & PR

- Conventional Commits: `feat(<name>): add <Game Title> game`.
- Never `git add -A` / `git add .` / `git checkout <branch> -- .` — this worktree has
  local-only files; stage explicit paths only.
- Fetch/merge latest `develop` before branching and pushing.
- PR body must include `Closes #<issue#>` when an issue exists.
