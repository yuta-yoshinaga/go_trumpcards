# frontend/ -- React Frontend Rules

This directory contains the React frontend (Vite + React + TypeScript).

## Package Manager Rule

**Always use `bun` instead of `npm`, and `bunx` instead of `npx`.** This project uses Bun as the sole JavaScript package manager and script runner.

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
| Pages | `src/pages/*.test.tsx` | On-mount API calls, rendering for each game phase/state, button interactions |

**Branch coverage (C1) must be 100%** for `src/api`, `src/components`, `src/pages`, and `src/utils`.

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
- **Avoid cumulative timeouts** -- use `waitFor` with reasonable timeouts; don't chain multiple long waits
- **Use `.first()` on `.or()` chains** -- Playwright strict mode requires a single element; use `.first()` when combining locators
- **Scope selectors carefully** -- e.g., scope card selectors to exclude NavBar elements to avoid false matches
- **Handle confirm dialogs** -- if the game has a reset confirmation dialog, click it in the test after reset

## i18n (Internationalization)

The Web GUI supports Japanese (ja) and English (en) via **react-i18next** with **i18next-browser-languagedetector**.

- **Config**: `src/i18n/index.ts`
- **Translation files**: `src/i18n/locales/{ja,en}/{common,blackjack,poker,oldmaid,daifugo,sevens,doubt,holdem,hearts,memory,klondike,freecell,baccarat,spades}.json`
- **In components**: use the `useTranslation()` hook
- **In non-component files** (e.g., `playerUtils.ts`, `messages.ts`, `gameConstants.ts`): import the `i18n` instance directly
- **Tests**: i18n is initialized in `src/test/setup.ts` with ja translations loaded
- **Server responses**: Web presenters send `messageCode` and `messageParams` alongside `message` for i18n-ready frontend rendering

## Pre-commit checks

```sh
bun run build        # Build React app
bun run check        # Biome lint + format check
bun run test         # Run Vitest unit tests
```
