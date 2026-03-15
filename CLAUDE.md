# CLAUDE.md

Go trump card game algorithms -- BlackJack, Poker, Old Maid, Daifugo, Sevens, Doubt, Texas Hold'em, Hearts, Memory, Klondike, Baccarat. Clean Architecture with CLI and Web GUI (React + Go REST API).

## Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [npm](https://www.npmjs.com/) | 11.x |

## Commands

```sh
# Run games
go run ./cmd/trumpcards                    # Interactive mode (switch games with 'switch <game>')
go run ./cmd/trumpcards --lang en          # Interactive mode in English
go run ./cmd/trumpcards blackjack          # BlackJack CLI
go run ./cmd/trumpcards --lang en blackjack  # BlackJack CLI in English
go run ./cmd/trumpcards poker      # 5-card Draw Poker CLI
go run ./cmd/trumpcards oldmaid    # Old Maid CLI
go run ./cmd/trumpcards daifugo    # Daifugo CLI
go run ./cmd/trumpcards sevens     # Sevens (7並べ) CLI
go run ./cmd/trumpcards doubt      # Doubt (ダウト) CLI
go run ./cmd/trumpcards holdem     # Texas Hold'em CLI
go run ./cmd/trumpcards hearts     # Hearts CLI
go run ./cmd/trumpcards memory     # Memory (神経衰弱) CLI
go run ./cmd/trumpcards klondike   # Klondike (ソリティア) CLI
go run ./cmd/trumpcards baccarat   # Baccarat (バカラ) CLI
go run ./cmd/trumpcards update     # Self-update to the latest version
go run ./cmd/trumpcards web        # Start REST API + web GUI server (via CLI)
go run ./cmd/server         # Start REST API + web GUI server (direct)

# Test
go test -tags test ./...                                              # Run all Go tests
go test -tags test -coverprofile=coverage.out -covermode=atomic ./... # Coverage report

# Format
goimports -w ./...           # Format and organize imports (use goimports, not gofmt)

# Frontend
cd frontend && npm install   # Install dependencies
cd frontend && npm run build # Build React app
cd frontend && npm run check # Biome lint + format check
cd frontend && npm test      # Run Vitest unit tests
cd frontend && npm run e2e   # Run Playwright E2E tests

# Docker
docker build -t go_trumpcards .
docker run --rm -d -p 8080:8080 go_trumpcards
```

## Go Formatting Rule

**After editing any Go source file, always run `goimports -w` on the modified files before committing.** Use `goimports`, not `gofmt`.

## Architecture

Clean Architecture: `infrastructure` -> `adapter` -> `usecase` -> `domain`. See [`docs/architecture.md`](docs/architecture.md) for full details.

## Testing

**Unit tests are mandatory. Every implementation must ship with tests in the same commit.**

**All code changes must follow the TDD cycle (Red-Green-Refactor):**

1. **Red** -- Write a failing test first. Before writing any production code, create or modify a test that captures the expected behavior and confirm it fails.
2. **Green** -- Write the minimum code to make the failing test pass. Do not add extra functionality beyond what the test requires.
3. **Refactor** -- Clean up (naming, structure, duplication) without changing behavior. Verify all tests still pass.

Apply this cycle at every layer: Domain, Use cases, Presenters, Controllers.

**Self-review checklist** -- before marking any task complete:

1. All Go tests pass: `go test -tags test ./...`
2. Go files formatted: `goimports -w` on modified files
3. Frontend checks pass (if applicable): `cd frontend && npm run build && npm run check && npm test`
4. Branch coverage is 100% for modified packages (excluding `cmd/` and `internal/infrastructure/`)

Also see:
- [`internal/CLAUDE.md`](internal/CLAUDE.md) -- Go-specific testing rules
- [`frontend/CLAUDE.md`](frontend/CLAUDE.md) -- Frontend-specific testing rules

### Coverage standard

The `cmd/` and `internal/infrastructure/` directories are excluded from coverage requirements. For all other packages under `internal/`, **branch coverage (C1) must be 100%**.

When writing tests, always verify branch coverage--not just statement coverage (C0)--by ensuring every conditional branch (if/else, switch cases, loop exit conditions, etc.) is exercised.

### Coverage requirements

When adding or modifying any game logic, provide tests for all four layers:

| Layer | Location | What to test |
|-------|----------|--------------|
| Domain | `internal/domain/*_test.go` | All public methods, edge cases, boundary values |
| Use cases | `internal/usecase/*Interactor_test.go` | Each interactor method via a mock presenter |
| Presenters | `internal/adapter/presenter/*_test.go` | CUI text output and Web JSON output for every game phase |
| Controllers | `internal/adapter/controller/*_test.go` | Every supported command including unknown/empty input |

### Mock pattern

- **Presenter mocks**: `internal/usecase/presenter/*_mock.go` -- implement the presenter interface using `testify/mock`
- **Interactor mocks**: `internal/adapter/controller/usecase/*_mock.go` -- implement the interactor interface using `testify/mock`
- Follow the existing `BlackJack*_mock.go` files as the reference pattern

### Writing deterministic tests

Card games involve shuffling, so tests must not depend on random outcomes:

- **Avoid auto-hit/draw**: Give the dealer/CPU a score that prevents automatic card draws (e.g., BlackJack dealer >= 17, Poker dealer rank >= Two Pair so `dealerExchange` is skipped)
- **Force deterministic draws in OldMaid**: Give the target player exactly 1 card so `rand.Intn(1)` always returns 0
- **Never assert on shuffled deck order**: Set up game state manually via `AddCard` instead of relying on `Reset`/`Shuffle`

### Verifying tests

Always run the full test suite before committing and ensure it passes:

```sh
go test -tags test ./...
```

### Frontend testing

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

### E2E testing

E2E tests use **Playwright** (Chromium only) and live in `frontend/e2e/`. They verify game flows (navigation, button availability, phase transitions) against the real Go server.

```sh
cd frontend && npm run e2e          # Run E2E tests (auto-starts Go server on port 8080)
cd frontend && npm run e2e:headed   # Run E2E tests in headed browser
cd frontend && npm run e2e:ui       # Run with Playwright UI
```

E2E tests should not assert on specific card values (randomness). Instead, verify flow: button visibility, phase transitions, and reset behavior.

### i18n (Internationalization)

The Web GUI supports Japanese (ja) and English (en) via **react-i18next** with **i18next-browser-languagedetector**.

- **Config**: `frontend/src/i18n/index.ts`
- **Translation files**: `frontend/src/i18n/locales/{ja,en}/{common,blackjack,poker,oldmaid,daifugo,sevens,doubt,holdem,hearts,memory,klondike,baccarat}.json`
- **In components**: use the `useTranslation()` hook
- **In non-component files** (e.g., `playerUtils.ts`, `messages.ts`, `gameConstants.ts`): import the `i18n` instance directly
- **Tests**: i18n is initialized in `frontend/src/test/setup.ts` with ja translations loaded
- **Server responses**: Web presenters send `messageCode` and `messageParams` alongside `message` for i18n-ready frontend rendering

## Documentation Maintenance

**When making code changes, always update the following documentation in the same commit:**

| Change type | Documents to update |
|-------------|---------------------|
| Add/remove a game | [`README.md`](README.md) (Description, Run section), [`CLAUDE.md`](CLAUDE.md) (Commands), [`docs/games.md`](docs/games.md) |
| Add/remove a CLI command (`cmd/cli/main.go`) | [`README.md`](README.md) (Run section), [`CLAUDE.md`](CLAUDE.md) (Commands) |
| Add/remove a Web API endpoint | [`docs/architecture.md`](docs/architecture.md) (Web API in Key patterns), [`api/openapi.yaml`](api/openapi.yaml) |
| Change request/response schema of a Web API endpoint | [`api/openapi.yaml`](api/openapi.yaml) |
| Change architecture or layer structure | [`README.md`](README.md) (Architecture), [`CLAUDE.md`](CLAUDE.md) (Architecture), [`docs/architecture.md`](docs/architecture.md) |
| Change Git workflow or CI/CD | [`CLAUDE.md`](CLAUDE.md) (Git Workflow) |
| Modify anything under `frontend/` | Run `cd frontend && npm run build`, `cd frontend && npm run check`, and `cd frontend && npm test` and ensure all three pass before committing |
| Add/remove frontend source files or change testing approach | Update Testing section in [`CLAUDE.md`](CLAUDE.md) (Frontend testing) and [`frontend/CLAUDE.md`](frontend/CLAUDE.md) |
| Change frontend tooling or scripts | [`frontend/README.md`](frontend/README.md) (Scripts, Tooling) |
| Change game rules or game flow logic | `docs/manual/cui/<game>.md` and `docs/manual/web/<game>.md` for the affected game |
| Change Go testing policy or mock patterns | Update Testing section in [`CLAUDE.md`](CLAUDE.md) and [`internal/CLAUDE.md`](internal/CLAUDE.md) |

Use commit type `docs` (or include doc changes in the same commit as the code change) following the Conventional Commits format.

### Intermediate design docs

**Do NOT commit intermediate design documents (e.g., `docs/superpowers/specs/`) to the repository.** These documents are not maintained after implementation and become tech debt. Instead:

- **Design specs and brainstorming output**: Post as a comment on the relevant GitHub issue
- **Architecture Decision Records (ADRs)**: These ARE worth committing to `docs/adr/` — they capture the *why* behind decisions and remain valuable long-term

## Git Workflow

- **`develop`**: Default branch; target for all PRs. CodeQL analysis and `golangci-lint` run on push/PR.
- **`master`**: Triggers automatic version bump, git tag, and GitHub Release.
- **PR Summary**: When creating a PR, if there is an associated issue, the PR description must explicitly close the issue (e.g., `Closes #123`).

## Commit Message Format

All commit messages must follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`

**Rules:**
- Lowercase description, no trailing period
- Imperative mood (e.g., "add feature" not "added feature")
- Breaking changes: `BREAKING CHANGE:` in footer or `!` after type/scope

## Workflow & Principles

### Plan Node Default

- Enter plan mode for ANY non-trivial task (3+ steps or architectural decisions)
- If something goes sideways, STOP and re-plan immediately -- don't keep pushing
- Use plan mode for verification steps, not just building
- Write detailed specs upfront to reduce ambiguity

### Subagent Strategy

- Use subagents liberally to keep main context window clean
- Offload research, exploration, and parallel analysis to subagents
- For complex problems, throw more compute at it via subagents
- One task per subagent for focused execution

### Self-Improvement Loop

- After ANY correction from the user: update `tasks/lessons.md` (create on demand; not committed) with the pattern
- Write rules for yourself that prevent the same mistake
- Ruthlessly iterate on these lessons until mistake rate drops
- Review lessons at session start for relevant project

### Verification Before Done

- Never mark a task complete without proving it works
- Diff behavior between main and your changes when relevant
- Ask yourself: "Would a staff engineer approve this?"
- Run tests, check logs, demonstrate correctness

### Demand Elegance (Balanced)

- For non-trivial changes: pause and ask "is there a more elegant way?"
- If a fix feels hacky: "Knowing everything I know now, implement the elegant solution"
- Skip this for simple, obvious fixes -- don't over-engineer
- Challenge your own work before presenting it

### Autonomous Bug Fixing

- When given a bug report: just fix it. Don't ask for hand-holding
- Point at logs, errors, failing tests -- then resolve them
- Zero context switching required from the user
- Go fix failing CI tests without being told how

### Task Management

1. **Plan First**: Write plan to `tasks/todo.md` (create on demand; not committed) with checkable items
2. **Verify Plan**: Check in before starting implementation
3. **Track Progress**: Mark items complete as you go
4. **Explain Changes**: High-level summary at each step
5. **Document Results**: Add review section to `tasks/todo.md`
6. **Capture Lessons**: Update `tasks/lessons.md` after corrections

### Core Principles

- **Simplicity First**: Make every change as simple as possible. Impact minimal code.
- **No Laziness**: Find root causes. No temporary fixes. Senior developer standards.
- **Minimal Impact**: Changes should only touch what's necessary. Avoid introducing bugs.
- **Dead Code Cleanup**: When modifying code, always remove any dead code or dead files you encounter. Use `golang.org/x/tools/cmd/deadcode` for Go and `knip` for TypeScript to identify unused code. Verify findings manually before deleting -- static analysis tools can produce false positives (e.g., interface implementations called via reflection, mock methods). Delete confirmed dead code in the same commit as your feature or fix.

## Detailed Context

| Topic | File |
|-------|------|
| Architecture & key patterns | [`docs/architecture.md`](docs/architecture.md) |
| Game descriptions & entities | [`docs/games.md`](docs/games.md) |
| Go backend rules | [`internal/CLAUDE.md`](internal/CLAUDE.md) |
| Frontend rules | [`frontend/CLAUDE.md`](frontend/CLAUDE.md) |
