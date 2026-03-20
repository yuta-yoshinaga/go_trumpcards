# React + TypeScript + Vite

This project uses [Vite](https://vitejs.dev/) with [React](https://react.dev/) and [TypeScript](https://www.typescriptlang.org/).

## Scripts

- `bun run dev`: Start the development server
- `bun run build`: Build for production
- `bun run preview`: Preview the production build
- `bun run check`: Run [Biome](https://biomejs.dev/) to check for linting and formatting errors
- `bun run check:write`: Run Biome and automatically fix linting/formatting errors
- `bun run test`: Run tests with [Vitest](https://vitest.dev/)
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
