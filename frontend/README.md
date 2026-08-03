# React + TypeScript + Vite

This project uses [Vite](https://vitejs.dev/) with [React](https://react.dev/) and [TypeScript](https://www.typescriptlang.org/).

## Scripts

- `bun run dev`: Start the development server
- `bun run build`: Build for production
- `bun run preview`: Preview the production build
- `bun run lint`: Run [Biome](https://biomejs.dev/) lint checks only
- `bun run lint:tokens`: Verify Tailwind classes use design tokens (no hard-coded values)
- `bun run format`: Run Biome formatter and auto-write fixes
- `bun run check`: Run Biome (lint + format) **and every guard script** — design tokens, a11y, i18n coverage, dependency licenses, and more. The authoritative list is the `check` entry in `package.json`; each guard prints its own OK line, so read that output rather than a list here. Prefer this over `bun run lint`: the guards catch most regressions, not Biome
- `bun run check:write`: Run Biome and automatically fix linting/formatting errors
- `bun run test`: Run tests with [Vitest](https://vitest.dev/)
- `bun run test:watch`: Run Vitest in watch mode
- `bun run test:coverage`: Run tests with coverage
- `bun run e2e`: Run [Playwright](https://playwright.dev/) E2E tests (auto-starts Go server)
- `bun run e2e:ui`: Run Playwright E2E tests with UI
- `bun run e2e:headed`: Run Playwright E2E tests in headed mode

- `bun run docs:generate`: Generate [TypeDoc](https://typedoc.org/) API documentation into `docs/`

## Tooling

- **Build**: Vite
- **Framework**: React
- **Language**: TypeScript
- **Server State**: TanStack React Query
- **Linting & Formatting**: Biome
- **Unit Testing**: Vitest + React Testing Library
- **E2E Testing**: Playwright
- **Styling**: Tailwind CSS (v4)
- **API Docs**: TypeDoc
