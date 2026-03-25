# New Game Addition Checklist

When adding a new game, follow this checklist to avoid post-feat fix commits. Complete ALL items before creating the PR.

## Backend (Go)

1. **Domain**: Create `internal/domain/<Game>.go`, `<Game>Player.go`, `<Game>Config.go` (if configurable)
2. **Reuse shared helpers**: `deal_helper.go` (dealAllCards), `hand_eval.go` (hand evaluation), `betting.go` (chip/betting), `play_style_helper.go` (CPU styles), `player_helpers.go` (resetPlayers), `hesitation.go` (CPU delay), `memory_manager.go`/`memory_decay.go` (CPU memory AI), `GamePlayer.go` (base player struct), `ChipHolder.go` (chip system), `kicker.go` (kicker comparison)
3. **Interactor**: `internal/usecase/<Game>Interactor.go` with presenter interface in `internal/usecase/presenter/`
4. **Controller**: CUI controller in `internal/adapter/controller/`, Web controller in `internal/adapter/controller/`, reuse `cuiutil` package for input parsing and `ClampIntPtr` for config validation
5. **Presenter**: CUI and Web presenters in `internal/adapter/presenter/`, reuse `buildCuiOutput`, `cuiCardListStr`, `ActionLogOutput` helpers, `WebOutputBase` for common web output fields
6. **Infrastructure**: Register in `cmd/trumpcards/main.go` (CLI) and `internal/infrastructure/web/TrumpCardsWeb.go` (API route)
7. **Run `goimports -w` and `golangci-lint run ./...`** on all new files
8. **80%+ branch coverage** for all new packages

## Frontend (React)

9. **Page**: `frontend/src/pages/<Game>Page.tsx` with test file, reuse `useGamePageSetup` hook, `usePhaseNames`, `gameReplay`, `useCardDimensions`, `gameExec` API helper
10. **Shared components**: Use `PhaseIndicator`, `SettingsPanel`, `ConfirmDialog`, `ActionLogSection`, `GameFooter`, `GameMessageBox`, `AnimatedCardBack`, `ErrorBoundary`
11. **i18n**: Add `frontend/src/i18n/locales/{ja,en}/<game>.json` translation files
12. **Router**: Add route in `frontend/src/App.tsx` and NavBar entry
13. **Run `bun run build && bun run check && bun run test`**

## Documentation (same commit)

14. **`README.md`**: Add game description and CLI command
15. **`CLAUDE.md`**: Add game name to available games list in Commands section
16. **`docs/games.md`**: Add game entity description
17. **`docs/architecture.md`**: Update endpoint count and list
18. **`api/openapi.yaml`**: Add endpoint path, tag definition, and request/response schemas in components
19. **`docs/manual/cui/<game>.md`** and **`docs/manual/web/<game>.md`**: Add game manuals

## Final verification

20. `go test -tags test ./...` -- all tests pass
21. `golangci-lint run ./...` -- no warnings
22. `cd frontend && bun run build && bun run check && bun run test` -- all pass
23. **E2E test**: Create `frontend/e2e/<game>.spec.ts` with basic game flow test
24. `cd frontend && bun run e2e` -- all E2E tests pass
