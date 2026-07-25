# frontend/ -- React Frontend Rules

This directory contains the React frontend (Vite + React + TypeScript).

## Design system

**Always read [`../DESIGN.md`](../DESIGN.md) before making any visual / UI changes.** Fonts, colors, spacing, motion, and the WCAG 2.5.5 AAA tap-target rule (44×44 px minimum for interactive controls) are defined there. Game background colors live in `src/styles/gameTheme.ts` (SSoT — every game page should reference its own key, not hardcode another game's theme).

## Standard page structure

Every game page (`src/pages/<X>Page.tsx`) follows the same skeleton — copy from an existing page like `SpadesPage.tsx` or `HeartsPage.tsx` rather than inventing a new structure. Key elements:

1. **Outer `XPage` exports the TutorialWrapper, inner `XPageContent` holds the implementation** — this split is required because `TutorialWrapper` provides the context that hooks inside `XPageContent` consume.
2. **Standard hooks**: `useGamePageSetup(gameName)` provides i18n + action-log + reset-confirm state; `useGameApi(apiFn, opts)` manages server state with loading / error / retry.
3. **Mount-time reset**: pages call the api command `"reset"` inside a `useEffect` on mount to fetch a fresh game from the server.
4. **Skeleton fallback**: render `<XSkeleton />` while `state` is `null`.
5. **`GamePageShell`** wraps most pages and provides the background, page heading, phase indicator, reset confirmation dialog, and win celebration overlay (#1650). Migration recipe for unmigrated pages:
   - Replace the outer `<div className="flex-1 flex flex-col min-h-0 ${theme.bg}" aria-busy={loading}>` block with `<GamePageShell>`.
   - Drop imports and JSX for `GamePageHeading`, `PhaseIndicator`, `TutorialButton`, `ManualButton`, `WinCelebration`, `GameResetDialog`, and `useGameRoundGuard` — the shell calls them all.
   - Pass `headerExtra` for any per-page header chips (e.g. `<CliToggle>`, chip / score readouts).
   - Pass `winShow` when the celebration condition is stricter than plain `gameEndFlag` (e.g. `humanWon` for two-player games where only the human's win is celebrated).
   - Pass `onCelebrate` to trigger sound effects or other side effects when the celebration starts.
   - Omit `isHumanTurn` for single-player solitaire pages with no "your turn / waiting" concept — the shell forwards `undefined` to `PhaseIndicator` and the turn label is hidden.

Shared building blocks:

| Hook / component | Purpose |
|------------------|---------|
| `useGamePageSetup(gameName)` | i18n (`t`, `tc`), action-log state, reset-confirm dialog state |
| `useGameApi(apiFn, opts)` | Server-state management with loading / error / retry |
| `GamePageShell` | Background + heading + phase indicator + reset dialog wrapper (most pages use this) |
| `TutorialWrapper` | Tutorial provider + i18n init (always wraps the exported `XPage`) |
| `useCliMode` + `useCliGame` + `<CliTerminal>` + `<CliToggle>` | CLI fallback mode, available on most pages. Per-game parsing/formatting lives in `src/utils/cli/commands/<game>Commands.ts` + `src/utils/cli/formatters/<game>Formatter.ts` |

## Package Manager Rule

**Always use `bun` instead of `npm`/`node`, and `bunx` instead of `npx`.** This project uses Bun as the sole JavaScript runtime, package manager, and script runner. Never invoke `node ./node_modules/...` directly — use `bun` or `bunx` instead.

## Testing

**TDD cycle (Red-Green-Refactor):**

1. **Red** -- Write a failing test (`bun run test -- --run TestName` confirms failure)
2. **Green** -- Write the minimum code to pass the test
3. **Refactor** -- Clean up while keeping all tests green (`bun run test`)

The test stack is **Vitest + React Testing Library + jest-dom**.

| Layer | Location | What to test |
|-------|----------|--------------|
| API client | `src/api/*.test.ts` | Correct URL, request body, and error handling for every API method |
| Components | `src/components/*.test.tsx` | Rendered output, props, event handlers |
| CLI components | `src/components/cli/*.test.tsx` | Terminal rendering, command input, toggle state |
| Pages | `src/pages/*.test.tsx` | On-mount API calls, rendering for each game phase/state, button interactions |
| Hooks | `src/hooks/*.test.ts` | State transitions, localStorage persistence, return values |
| CLI hooks | `src/hooks/useCliMode.test.ts`, `src/hooks/useCliGame.test.ts` | CLI mode toggle, log management, command orchestration |
| Utils | `src/utils/**/*.test.ts` | Pure function input/output, edge cases |
| CLI commands | `src/utils/cli/commands/*.test.ts` | Command parsing, alias mapping, error handling |
| CLI formatters | `src/utils/cli/formatters/*.test.ts` | Game state text formatting, edge cases |

**Branch coverage (C1) must be 80% or higher** for `src/api`, `src/components`, `src/pages`, and `src/utils`. Focus testing effort on business logic and critical paths rather than exhaustively covering every conditional branch.

### Patterns

- **Mock the API module**: use `vi.mock('../api/gameApi', ...)` inside page test files; access the typed mock with `vi.mocked(api.exec)`
- **Wrap router-dependent components**: render `NavBar` (and any component using `useLocation`) inside `<MemoryRouter initialEntries={['/path']}>`
- **Wait for async effects**: use `waitFor(() => expect(...))` after render when the component fires an API call in `useEffect`
- **Query buttons by role**: when a text string appears in multiple elements (e.g., "交換" appears on both cards and a button), use `screen.getByRole('button', { name: '交換' })` instead of `getByText`
- **Wrap with QueryClientProvider**: page tests and hook tests must render inside a `QueryClientProvider` (use `renderWithProviders` from `src/test/renderWithProviders.tsx`)

## E2E testing

E2E tests use **Playwright** (Chromium only) and live in `e2e/`. They verify game flows against the real Go server.

```sh
bun run e2e          # Run E2E tests (auto-starts Go server on port 8080)
bun run e2e:headed   # Run E2E tests in headed browser
bun run e2e:ui       # Run with Playwright UI
```

### E2E test guidelines (avoiding flaky tests)

- **Never assert on specific card values** -- card games involve shuffling; assertions on card content will be flaky
- **Verify phase transitions** -- check that buttons appear/disappear and the game progresses through phases
- **Verify reset behavior** -- ensure the game can be reset and restarted
- **Use timeout constants** -- import `TIMEOUT_QUICK`, `TIMEOUT_ACTION`, `TIMEOUT_TRANSITION`, `TIMEOUT_GAME_LOOP` from `./helpers` instead of magic numbers
- **Use `isVisibleWithin()` for conditional checks** -- import from `./helpers` instead of using `.isVisible({ timeout }).catch(() => false)` directly
- **Never use `.catch(() => false)` on bare `isVisible()`** -- Playwright's `locator.isVisible()` already returns `false` for missing elements without throwing
- **Avoid count-then-act patterns** -- use `locator.first().click({ timeout })` with try/catch instead of checking `.count()` then clicking, to prevent race conditions
- **Use `.first()` on `.or()` chains** -- Playwright strict mode requires a single element; use `.first()` when combining locators
- **Scope selectors carefully** -- e.g., scope card selectors to exclude NavBar elements to avoid false matches
- **Handle confirm dialogs** -- if the game has a reset confirmation dialog, click it in the test after reset

## i18n (Internationalization)

The Web GUI supports Japanese (ja) and English (en) via **react-i18next** with **i18next-browser-languagedetector**.

- **Config**: `src/i18n/index.ts`
- **Translation files**: `src/i18n/locales/{ja,en}/<game>.json` (each game name + `common.json`, `tutorial.json`)
- **In components**: use the `useTranslation()` hook
- **In non-component files** (e.g., `playerUtils.ts`, `messages.ts`, `gameConstants.ts`): import the `i18n` instance directly
- **Tests**: i18n is initialized in `src/test/setup.ts` with ja translations loaded
- **Server responses**: Web presenters send `messageCode` and `messageParams` alongside `message` for i18n-ready frontend rendering

## TSDoc Comments

All exported symbols (types, interfaces, functions, components, constants, hooks) must have TSDoc comments.

- **Comment style**: `/** Brief description */`
- **React components**: describe what the component renders
- **Hooks**: describe what the hook provides
- **API functions**: describe what API endpoint is called
- **Utility functions**: describe what the function does
- **Generated docs**: run `bun run docs:generate` to produce HTML documentation in `docs/` (gitignored)

## Pre-commit checks

```sh
bun run build && bun run check && bun run test
```

## Tutorial System

### Key components

| Component | Location | Purpose |
|-----------|----------|---------|
| `TutorialWrapper` | `src/components/tutorial/TutorialWrapper.tsx` | Combines TutorialProvider + i18n; wraps game page with `gameName` and `steps` props |
| `TutorialButton` | `src/components/tutorial/TutorialButton.tsx` | Shared tutorial start button |
| `TutorialProvider` | `src/providers/TutorialProvider.tsx` | Context provider; renders overlay (used internally by TutorialWrapper) |
| `TutorialOverlay` | `src/components/tutorial/TutorialOverlay.tsx` | Full-screen overlay with SVG mask spotlight |
| `useTutorial` | `src/hooks/useTutorial.ts` | State management (step progression, localStorage, resume/restart) |
| `useGameHint` | `src/hooks/useGameHint.ts` | Frontend hints; registry-driven via `hintFactories` (covers nearly all games; currently 193). The supported set is the `HintGameName` union (`keyof typeof hintFactories`) — add a game by registering its factory there |
| `TutorialSuggestDialog` | `src/components/tutorial/TutorialSuggestDialog.tsx` | First-visit dialog; controlled by `useFirstVisit` hook |
| `TutorialProgressPanel` | `src/components/tutorial/TutorialProgressPanel.tsx` | Progress overview in NavBar |

### Adding a tutorial to a new game

1. Define `TutorialStep[]` array with `target` (CSS selector using `data-tutorial` attributes), `messageKey`, `placement`, and `advanceOn`
2. Add `data-tutorial="<step-name>"` attributes to the game page's key UI elements
3. Add tutorial step text to `src/i18n/locales/{ja,en}/<game>.json` under a `tutorial` key
4. Wrap the page content with `<TutorialWrapper gameName="<game>" steps={steps}>` and import `TutorialButton` from `../components/tutorial/TutorialButton`
