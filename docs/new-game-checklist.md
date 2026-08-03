# New Game Addition Checklist

When adding a new game, follow this checklist to avoid post-feat fix commits. Complete ALL items before creating the PR.

## Before you pick the name

0. **Screen the display name, and every feature name inside it** (side bets, bonus
   names, bet names). Traditional and folk games are free; a name invented and
   sold by an identifiable company is not. `scripts/jpo-trademark-search.mjs`
   runs the Japanese check — **search the rights holder as well as the mark**,
   since `21+3` was found only from its owner's portfolio and never from the
   term itself. Record the outcome in [`TRADEMARKS.md`](../TRADEMARKS.md), which
   `bun run check` enforces. See that file for the full policy.

## Backend (Go)

1. **Domain**: Create `internal/domain/<Game>.go`, `<Game>Player.go`, `<Game>Config.go` (if configurable)
2. **Reuse shared helpers**: `deal_helper.go` (dealAllCards), `hand_eval.go` (hand evaluation), `betting.go` (chip/betting), `play_style_helper.go` (CPU styles), `player_helpers.go` (resetPlayers), `trick_helpers.go` (`TrickCard` type + `ResolveTrickWinner` for trick-taking games), `trick_cpu_helpers.go` (`filterByDesign`/`filterAbove`/`filterBelow`/`pickLowest`/`pickHighest` tactical sub-steps for trick-taking CPU AI), `hesitation.go` (CPU delay), `memory_manager.go`/`memory_decay.go` (CPU memory AI), `GamePlayer.go` (base player struct), `ChipHolder.go` (chip system), `kicker.go` (kicker comparison)
3. **Interactor**: `internal/usecase/<Game>Interactor.go` with presenter interface in `internal/usecase/presenter/`
4. **Controller**: CUI controller in `internal/adapter/controller/`, Web controller in `internal/adapter/controller/`, reuse `cuiutil` package for input parsing and `ClampIntPtr` for config validation
5. **Presenter**: CUI and Web presenters in `internal/adapter/presenter/`, reuse `buildCuiOutput`, `cuiCardListStr`, `ActionLogOutput` helpers, `WebOutputBase` for common web output fields
6. **Infrastructure**: Register in `cmd/trumpcards/main.go` (CLI) and add four entries for the new game:
   - **(a)** `{Name: "<name>", Category: Category…}` in the `registry` slice in `internal/infrastructure/games/registry.go`. `Category` pins the game to one Cloudflare Worker binary-size bucket (`casino`, `classic`, `solo`, `extra`, `extra2`, or `extra3` — ADR-0032 added the fourth, ADR-0036 the fifth and sixth). The bucket is **not** a genre taxonomy and none of them is *the* overflow bucket: **pick the worker with the most gzip headroom, measured rather than assumed.** `.claude/skills/rebucket-game/scripts/measure.sh` prints every worker's size; a recent CI `tinygo-build` log gives the same figures without a local TinyGo.
   - **(b)** `BindWebController("<name>", …)` in `internal/infrastructure/games/games_server.go` — wires the HTTP server factory.
   - **(c)** `games.RegisterKVGame("<name>", games.Category…, …)` in the matching sub-package under `internal/infrastructure/games/{casino,classic,solo,extra,extra2,extra3}/` — the per-category split is what keeps each Cloudflare Worker WASM binary under the 1 MB gzipped free-tier limit, so the sub-package must match the `Category` in (a). `RegisterKVGame` panics at init if they disagree (via the underlying `games.BindWorker`), and `TestWorkerRegistrationsCoverAllGames` (in `registry_worker_consistency_test.go`) parses each sub-package source with `go/parser` and fails the build if the registry and a category's `RegisterKVGame` calls disagree (per ADR-0031, option 3).
   - **(d)** `GameRegistryEntry` entry in the `gameRegistry` slice in `internal/infrastructure/ui/GameManager.go` — the CLI-side wiring. Always use the `cuiEntry` helper with a `CuiHelpSpec`. For the standard help template, fill in `TitleKey`/`CommandKeys`/`SettingKeys`; for hand-authored help that does not fit the scaffold, set `CuiHelpSpec.Body` to the full help lines.
6b. **Frontend worker URL**: verify `frontend/src/api/gameApi.ts` `workerUrl` maps the new game to the same worker name as its `Category`.
7. **Run `goimports -w` and `golangci-lint run ./...`** on all new files
8. **80%+ branch coverage** for all new packages

## Frontend (React)

9. **Page**: `frontend/src/pages/<Game>Page.tsx` with test file, reuse `useGamePageSetup` hook, `usePhaseNames`, `gameReplay`, `useCardDimensions`, `gameExec` API helper. Add `useGameRoundGuard(!!state && !state.gameEndFlag)` so accidental tab close / reload during a round triggers the browser leave warning (pages using `GamePageShell` get this for free; otherwise call the hook directly).
10. **Shared components**: Use `PhaseIndicator`, `SettingsPanel`, `ConfirmDialog`, `ActionLogSection`, `GameFooter`, `GameMessageBox`, `AnimatedCardBack`, `ErrorBoundary`. Add a `GameKey` entry in `frontend/src/styles/gameTheme.ts` and reference `gameTheme.<key>` (do not reuse another game's key; the strict union type catches missing entries at compile time).
11. **CLI mode**: Wire `useCliMode`, `useCliGame`, `CliToggle`, and `CliTerminal` in the page. At minimum add a stub config (`parseCommand` returns error, empty `helpText`). Place `CliToggle` inside `PhaseIndicator` and wrap GUI content with `{cliEnabled ? <CliTerminal .../> : <>{/* GUI */}</>}`
12. **i18n**: Add `frontend/src/i18n/locales/{ja,en}/<game>.json` translation files (include `tutorial` keys if tutorial is added)
13. **Router + concierge profile**: Add route in `frontend/src/App.tsx` and NavBar entry in `frontend/src/constants/gameRoutes.ts`. The `gameRoutes` entry must include the `profile: GameProfile` field (4 axes — see `frontend/src/constants/discoverAxes.ts` for option order). The TypeScript compiler will reject the file if `profile` is missing. Also add the per-game `discover.blurb.<page-kebab>` and `discover.stretch_blurb.<page-kebab>` entries in both `frontend/src/i18n/locales/{ja,en}/discover.json` so the AI Game Concierge `/discover` recommendations show real text for the new game.
14. **Hint system**: Create `frontend/src/utils/hints/<game>Hint.ts` returning `HintResult | null` (or `null` stub if the game has no decisional hint), and register it in `frontend/src/hooks/useGameHint.ts` `hintFactories`. Add the matching `<game>Hint.test.ts` with 80%+ branch coverage. Wire `useGameHint` + `HintToggle` into the page. **`bun run check` enforces the registration** (`frontend/scripts/check-hint-coverage.mjs`) — four of the five games before the guard existed had skipped this step, because nothing failed when they did.
15. **Tutorial**: Wrap `<Game>Page.tsx` content with `<TutorialWrapper gameName="<game>" steps={steps}>`, import `TutorialButton`, add `data-tutorial` attributes to key UI elements, and add `tutorial` keys to the i18n JSON files.
16. **Run `bun run build && bun run check && bun run test`**

## Documentation (same commit)

17. **`README.md`**: Add game description and CLI command
18. **`CLAUDE.md`**: Add game name to available games list in Commands section
19. **`docs/games.md`**: Add game entity description
20. **`docs/architecture.md`**: Update endpoint count and list
21. **`api/openapi.yaml`**: Add endpoint path, tag definition, and request/response schemas in components.
    Guarded by three tests in `internal/infrastructure/games`: `TestOpenAPIMatchesRegistry` (one
    `POST /<game>/exec` per registered game, no orphans), `TestOpenAPIHasNoDanglingSchemaRefs` (every `$ref`
    resolves), and `TestOpenAPIErrorResponseMatchesTheSuccessSchema` (a path's `400` documents the same schema
    as its `200` -- every endpoint returns the game's own payload on both branches, an error being a normal
    response carrying a `message`). The three check different things: that the path exists, that its refs point
    at something, and that they point at the RIGHT something.
    **The file is CRLF** -- anything that rewrites it must preserve the line endings or the diff becomes every line.
22. **`docs/manual/cui/<game>.md`** and **`docs/manual/web/<game>.md`**: Create from the templates below and **include every required section**:
    - CUI (`docs/manual/cui_template.md`): `## ゲーム概要` / `## 起動方法` / `## ルール` / `## ゲームの流れ` (Mermaid flowchart — **mandatory**) / `## コマンド一覧` / `## 画面の見方` / `## 遊び方のコツ`
    - Web (`docs/manual/web_template.md`): `## ゲーム概要` / `## 起動方法` / `## ルール` / `## ゲームの流れ` (Mermaid flowchart — **mandatory**) / `## 画面の操作方法` / `## 画面構成` / `## 遊び方のコツ`
    - Do not skip sections even if short; a brief stub is preferable to a missing section so future readers can rely on a consistent structure.
23. **`frontend/src/constants/manualTexts.ts`**: Import the **Web** manual and add route mapping entry
24. **`frontend/src/constants/cuiManualTexts.ts`**: Import the **CUI** manual and add route mapping entry (this is the manual displayed when CLI mode is active — easy to forget)

## Final verification

25. `go test -tags test ./...` -- all tests pass
26. `golangci-lint run ./...` -- no warnings
27. `cd frontend && bun run build && bun run check && bun run test` -- all pass
28. **E2E test**: Create `frontend/e2e/<game>.spec.ts` with basic game flow test
29. `cd frontend && bun run e2e` -- all E2E tests pass

## Self-audit (perform before marking PR ready)

Run through this cross-check for the new `<game>`:

- [ ] `docs/manual/cui/<game>.md` exists and contains every section listed in item 22 (including the Mermaid flowchart)
- [ ] `docs/manual/web/<game>.md` exists and contains every section listed in item 22 (including the Mermaid flowchart)
- [ ] `frontend/src/constants/manualTexts.ts` has both the `import` and the route mapping for `<game>`
- [ ] `frontend/src/constants/cuiManualTexts.ts` has both the `import` and the route mapping for `<game>`
- [ ] `frontend/src/utils/hints/<game>Hint.ts` exists (real implementation or documented `null` stub)
- [ ] `frontend/src/hooks/useGameHint.ts` registers `<game>` in `hintFactories`
- [ ] `<Game>Page.tsx` is wrapped in `<TutorialWrapper>` and surfaces `TutorialButton` + `HintToggle`
- [ ] `internal/infrastructure/games/registry.go` has a `{Name, Category}` entry for `<game>`
- [ ] `internal/infrastructure/games/games_server.go` has a matching `BindWebController("<game>", ...)` call
- [ ] `internal/infrastructure/games/<category>/<category>.go` has a matching `games.RegisterKVGame("<game>", games.Category…, ...)` call in the correct category sub-package
- [ ] `internal/infrastructure/ui/GameManager.go` `gameRegistry` has a matching `GameRegistryEntry` for `<game>` (CLI wiring)
- [ ] `frontend/src/api/gameApi.ts` `workerUrl` maps `<game>` to the worker matching that `Category`
- [ ] Hardcoded game-count assertions are bumped: the `expected<Category>` consts in `internal/infrastructure/games/registry_test.go`, and the tutorial-progress counts in `frontend/src/hooks/useTutorialProgress.test.ts` + `frontend/src/components/tutorial/TutorialProgressPanel.test.tsx`
