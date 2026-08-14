---
globs: ["frontend/**/*.ts", "frontend/**/*.tsx"]
---

# Frontend (TypeScript/TSX) File Editing Rules

## Package Manager

**Always use `bun` instead of `npm`, and `bunx` instead of `npx`.** This project uses Bun as the sole JavaScript package manager and script runner.

## Pre-commit Checks (mandatory, all must pass)

```sh
cd frontend && bun run build   # React build
cd frontend && bun run check   # Biome lint + format check
cd frontend && bun run typecheck # TypeScript 7 type check (never bare `tsc`)
cd frontend && bun run test    # Vitest unit tests
```

## Testing

**Unit tests are mandatory.** Include them in the same commit as the implementation. Test stack: **Vitest + React Testing Library + jest-dom**

### TDD Cycle (Red → Green → Refactor)

Always follow this cycle before implementing:

1. **Red** — Write a failing test first. Create a test that captures the expected behavior before writing implementation code:
   ```sh
   cd frontend && bunx vitest run <file> -t "<test name>"  # Fails (Red)
   ```
2. **Green** — Write the minimum code to pass the test. Do not add extra functionality:
   ```sh
   cd frontend && bunx vitest run <file> -t "<test name>"  # Passes (Green)
   ```
3. **Refactor** — Clean up code while keeping tests green. Improve naming, structure, and remove duplication:
   ```sh
   cd frontend && bun run test  # All tests pass (after Refactor)
   ```

### Coverage Standard

**Branch coverage (C1) of 80% or higher** is required for the following 4 directories:

- `frontend/src/api`
- `frontend/src/components`
- `frontend/src/pages`
- `frontend/src/utils`

Focus on business logic and critical paths. Forced coverage of unreachable branches is unnecessary.

### Test Locations (by layer)

| Layer | Test file | What to test |
|-------|-----------|-------------|
| API client | `src/api/*.test.ts` | URL, request body, error handling |
| Components | `src/components/*.test.tsx` | Rendering, props, event handlers |
| Pages | `src/pages/*.test.tsx` | On-mount API calls, phase-specific rendering, button interactions |

### Test Patterns

- **API mocks**: Mock the API module with `vi.mock('../api/gameApi', ...)`; access via `vi.mocked(api.exec)`
- **Router-dependent components**: Wrap components using `useLocation` (e.g., `NavBar`) in `<MemoryRouter initialEntries={['/path']}>`
- **Async effect waiting**: Use `waitFor(() => expect(...))` for components that call APIs in `useEffect`
- **Button queries**: When text appears in multiple elements, use `screen.getByRole('button', { name: '...' })`
- **QueryClientProvider wrapping**: Page tests and hook tests must use `renderWithProviders` (`frontend/src/test/renderWithProviders.tsx`)

## i18n (Internationalization)

The Web GUI supports Japanese (ja) / English (en) via `react-i18next` + `i18next-browser-languagedetector`.

- **Config**: `frontend/src/i18n/index.ts`
- **Translation files**: `frontend/src/i18n/locales/{ja,en}/<game>.json`
- **In components**: use the `useTranslation()` hook
- **In non-component files** (e.g., `playerUtils.ts`): import the `i18n` instance directly
- **Tests**: ja translations are initialized in `frontend/src/test/setup.ts`

## Dead Code

- Always remove dead code encountered when modifying code
- Detection tool: `knip`
- Verify manually before deleting (beware of false positives from static analysis)
