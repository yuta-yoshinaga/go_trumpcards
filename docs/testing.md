# Testing Policy

**Unit tests are mandatory. Every implementation must ship with tests in the same commit.**

## TDD (Test-Driven Development)

**All code changes must follow the TDD cycle (Red-Green-Refactor).** See [`.ai/skills/tdd-flow.md`](../.ai/skills/tdd-flow.md) for the detailed workflow.

## Coverage standard

The `cmd/` and `internal/infrastructure/` directories are excluded from coverage requirements. For all other packages under `internal/`, **branch coverage (C1) must be 100%**.

When writing tests, always verify branch coverage--not just statement coverage (C0)--by ensuring every conditional branch (if/else, switch cases, loop exit conditions, etc.) is exercised.

## Coverage requirements

When adding or modifying any game logic, provide tests for all four layers:

| Layer | Location | What to test |
|-------|----------|--------------|
| Domain | `internal/domain/*_test.go` | All public methods, edge cases, boundary values |
| Use cases | `internal/usecase/*Interactor_test.go` | Each interactor method via a mock presenter |
| Presenters | `internal/adapter/presenter/*_test.go` | CUI text output and Web JSON output for every game phase |
| Controllers | `internal/adapter/controller/*_test.go` | Every supported command including unknown/empty input |

## Mock pattern

- **Presenter mocks**: `internal/usecase/presenter/*_mock.go` -- implement the presenter interface using `testify/mock`
- **Interactor mocks**: `internal/adapter/controller/usecase/*_mock.go` -- implement the interactor interface using `testify/mock`
- Follow the existing `BlackJack*_mock.go` files as the reference pattern

## Writing deterministic tests

Card games involve shuffling, so tests must not depend on random outcomes:

- **Avoid auto-hit/draw**: Give the dealer/CPU a score that prevents automatic card draws (e.g., BlackJack dealer >= 17, Poker dealer rank >= Two Pair so `dealerExchange` is skipped)
- **Force deterministic draws in OldMaid**: Give the target player exactly 1 card so `rand.Intn(1)` always returns 0
- **Never assert on shuffled deck order**: Set up game state manually via `AddCard` instead of relying on `Reset`/`Shuffle`

## Verifying tests

Always run the full test suite before committing and ensure it passes:

```sh
go test -tags test ./...
```

## Frontend testing

Frontend unit tests are also mandatory. The test stack is **Vitest + React Testing Library + jest-dom**.

| Layer | Location | What to test |
|-------|----------|--------------|
| API client | `frontend/src/api/*.test.ts` | Correct URL, request body, and error handling for every API method |
| Components | `frontend/src/components/*.test.tsx` | Rendered output, props, event handlers |
| Pages | `frontend/src/pages/*.test.tsx` | On-mount API calls, rendering for each game phase/state, button interactions |

**Branch coverage (C1) must be 100%** for the four directories `frontend/src/api`, `frontend/src/components`, `frontend/src/pages`, and `frontend/src/utils`. When writing tests, always verify branch coverage--not just statement coverage (C0)--by ensuring every conditional branch (if/else, ternary, `??`, `&&`/`||` short-circuits, switch cases) is exercised.

**Patterns:**

- **Mock the API module**: use `vi.mock('../api/gameApi', ...)` inside page test files; access the typed mock with `vi.mocked(api.exec)`
- **Wrap router-dependent components**: render `NavBar` (and any component using `useLocation`) inside `<MemoryRouter initialEntries={['/path']}>`
- **Wait for async effects**: use `waitFor(() => expect(...))` after render when the component fires an API call in `useEffect`
- **Query buttons by role**: when a text string appears in multiple elements (e.g., "交換" appears on both cards and a button), use `screen.getByRole('button', { name: '交換' })` instead of `getByText`
- **Wrap with QueryClientProvider**: page tests and hook tests must render inside a `QueryClientProvider` (use `renderWithProviders` from `frontend/src/test/renderWithProviders.tsx`)

**Run build, Biome check, and frontend tests before committing:**

```sh
cd frontend && npm run build
cd frontend && npm run check
cd frontend && npm test
```

## E2E testing

E2E tests use **Playwright** (Chromium only) and live in `frontend/e2e/`. They verify game flows (navigation, button availability, phase transitions) against the real Go server.

```sh
cd frontend && npm run e2e          # Run E2E tests (auto-starts Go server on port 8080)
cd frontend && npm run e2e:headed   # Run E2E tests in headed browser
cd frontend && npm run e2e:ui       # Run with Playwright UI
```

E2E tests should not assert on specific card values (randomness). Instead, verify flow: button visibility, phase transitions, and reset behavior.

## i18n (Internationalization)

The Web GUI supports Japanese (ja) and English (en) via **react-i18next** with **i18next-browser-languagedetector**.

- **Config**: `frontend/src/i18n/index.ts`
- **Translation files**: `frontend/src/i18n/locales/{ja,en}/{common,blackjack,poker,oldmaid,daifugo,sevens,doubt,holdem}.json`
- **In components**: use the `useTranslation()` hook
- **In non-component files** (e.g., `playerUtils.ts`, `messages.ts`, `gameConstants.ts`): import the `i18n` instance directly
- **Tests**: i18n is initialized in `frontend/src/test/setup.ts` with ja translations loaded
- **Server responses**: Web presenters send `messageCode` and `messageParams` alongside `message` for i18n-ready frontend rendering
