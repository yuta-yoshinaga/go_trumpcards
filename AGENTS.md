# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Repository Overview

This repository contains a Go implementation of trump card game algorithms. The project is structured following the principles of Clean Architecture. The following games are implemented:

- **BlackJack**: CLI and Web GUI (chip/betting system, split, double down, insurance)
- **Poker (5-card Draw)**: CLI and Web GUI (chip/betting system)
- **Old Maid (Babanuki)**: CLI and Web GUI
- **Daifugo**: CLI and Web GUI
- **Sevens (7並べ)**: CLI and Web GUI (optional rules: tunnel, joker, CPU strategy)

## Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [npm](https://www.npmjs.com/) | 11.x |

## Common Commands

**Run the application:**
```sh
go run ./cmd/cli blackjack  # BlackJack CLI
go run ./cmd/cli poker      # 5-card Draw Poker CLI
go run ./cmd/cli oldmaid    # Old Maid CLI
go run ./cmd/cli daifugo    # Daifugo CLI
go run ./cmd/cli sevens     # Sevens (7並べ) CLI
go run ./cmd/cli web        # Start REST API + web GUI server (via CLI)
go run ./cmd/server         # Start REST API + web GUI server (direct)
```

**Run tests:**
```sh
go test ./...                                              # Run all tests
go test ./internal/domain/...                              # Run tests in a specific package
go test ./internal/domain/ -run TestBlackJack              # Run a single test by name
go test -coverprofile=coverage.out -covermode=atomic ./... # Run all tests with coverage report
go tool cover -func=coverage.out                           # Show coverage summary by function
go tool cover -html=coverage.out -o coverage.html          # Generate HTML coverage report
```

**Manage dependencies:**
```sh
go mod tidy
```

**Docker:**
```sh
docker build -t go_trumpcards .           # Build Docker image
docker run --rm -d -p 8080:8080 go_trumpcards  # Run Docker container
```

**Build frontend (React):**
```sh
cd frontend && npm install && npm run build   # Build React app to public/ (for local testing with Go server; Docker handles this automatically)
cd frontend && npm run check                  # Run Biome lint + format check
cd frontend && npm run check:write            # Run Biome lint + format check and auto-fix
cd frontend && npm test                       # Run frontend unit tests (Vitest)
cd frontend && npm run test:coverage          # Run frontend tests with coverage (outputs to frontend/coverage/)
```

## Architecture

The project implements Clean Architecture with strict layer dependency rules (outer depends on inner, never the reverse):

```
cmd/
  cli/                         # CLI entrypoint (all games + web server)
  server/                      # Web server dedicated entrypoint
internal/
  domain/                      # Core business logic (innermost)
  usecase/                     # Application business rules (interactors)
    presenter/                 # Presenter interfaces defined here
  adapter/                     # Convert data between layers
    controller/                # Route commands to use cases
    presenter/                 # Implement presenter interfaces for CUI and Web
  infrastructure/              # Outermost layer
    ui/                        # CLI runner
    web/                       # REST API server (go-json-rest)
api/                           # OpenAPI specification
frontend/                      # React frontend source (Vite + React + TypeScript)
  src/
    api/                       # API client functions (fetch wrappers)
    components/                # Shared React components (NavBar, CardImage, CardBack)
    pages/                     # Game page components (BlackJackPage, PokerPage, OldMaidPage, DaifugoPage)
    types/                     # TypeScript type definitions for card/game data
public/                        # Built frontend assets served by Go web server
  assets/                      # Vite-compiled JS/CSS bundles
  images/                      # Card images (PNG)
  css/                         # Bootstrap CSS
```

**Data flow:** `infrastructure` → `adapter` → `usecase` → `domain`

### Key patterns

- **Presenter pattern**: `internal/usecase/presenter/` defines output interfaces (e.g., `BlackJackPresenter`). `internal/adapter/presenter/` provides concrete implementations (CUI vs Web). Presenters are injected into interactors.
- **Mock presenters**: `*_mock.go` files in `internal/adapter/presenter/` and `internal/usecase/presenter/` are used in tests to avoid I/O.
- **Web API**: Five endpoints — `POST /blackjack/exec` (BlackJack), `POST /poker/exec` (Poker), `POST /oldmaid/exec` (Old Maid), `POST /daifugo/exec` (Daifugo), and `POST /sevens/exec` (Sevens) — accept JSON with a `Cmd` field and game state.

### Games implemented

- **BlackJack**: Entities in `internal/domain/BlackJack.go`, `internal/domain/BlackJackPlayer.go`, `internal/domain/BlackJackHand.go`; interactor in `internal/usecase/BlackJackInteractor.go`. Features chip/betting system, split, double down, insurance, and natural BJ 3:2 payout
- **Poker (5-card Draw)**: Entities in `internal/domain/Poker.go`, `internal/domain/PokerPlayer.go`; interactor in `internal/usecase/PokerInteractor.go`. Features chip/betting system with ante, bet/call/raise/check/fold, CPU betting AI, and improved CPU exchange strategy (flush/straight draw awareness)
- **Old Maid (Babanuki)**: Entities in `internal/domain/OldMaid.go`, `internal/domain/OldMaidPlayer.go`; interactor in `internal/usecase/OldMaidInteractor.go`
- **Daifugo**: Entities in `internal/domain/Daifugo.go`, `internal/domain/DaifugoPlayer.go`; interactor in `internal/usecase/DaifugoInteractor.go`
- **Sevens (7並べ)**: Entities in `internal/domain/Sevens.go`, `internal/domain/SevensPlayer.go`, `internal/domain/SevensConfig.go`; interactor in `internal/usecase/SevensInteractor.go`. Supports optional rules: tunnel (A↔K circular), joker, and CPU strategy

## Testing Policy

**Unit tests are mandatory. Every implementation must ship with tests in the same commit.**

### Coverage requirements

When adding or modifying any game logic, provide tests for all four layers:

| Layer | Location | What to test |
|-------|----------|--------------|
| Domain | `internal/domain/*_test.go` | All public methods, edge cases, boundary values |
| Use cases | `internal/usecase/*Interactor_test.go` | Each interactor method via a mock presenter |
| Presenters | `internal/adapter/presenter/*_test.go` | CUI text output and Web JSON output for every game phase |
| Controllers | `internal/adapter/controller/*_test.go` | Every supported command including unknown/empty input |

### Mock pattern

- **Presenter mocks**: `internal/usecase/presenter/*_mock.go` — implement the presenter interface using `testify/mock`
- **Interactor mocks**: `internal/adapter/controller/usecase/*_mock.go` — implement the interactor interface using `testify/mock`
- Follow the existing `BlackJack*_mock.go` files as the reference pattern

### Writing deterministic tests

Card games involve shuffling, so tests must not depend on random outcomes:

- **Avoid auto-hit/draw**: Give the dealer/CPU a score that prevents automatic card draws (e.g., BlackJack dealer ≥ 17, Poker dealer rank ≥ Two Pair so `dealerExchange` is skipped)
- **Force deterministic draws in OldMaid**: Give the target player exactly 1 card so `rand.Intn(1)` always returns 0
- **Never assert on shuffled deck order**: Set up game state manually via `AddCard` instead of relying on `Reset`/`Shuffle`

### Verifying tests

Always run the full test suite before committing and ensure it passes:

```sh
go test ./...
```

### Frontend testing

Frontend unit tests are also mandatory. The test stack is **Vitest + React Testing Library + jest-dom**.

| Layer | Location | What to test |
|-------|----------|--------------|
| API client | `frontend/src/api/*.test.ts` | Correct URL, request body, and error handling for every API method |
| Components | `frontend/src/components/*.test.tsx` | Rendered output, props, event handlers |
| Pages | `frontend/src/pages/*.test.tsx` | On-mount API calls, rendering for each game phase/state, button interactions |

**Patterns:**

- **Mock the API module**: use `vi.mock('../api/gameApi', ...)` inside page test files; access the typed mock with `vi.mocked(api.exec)`
- **Wrap router-dependent components**: render `NavBar` (and any component using `useLocation`) inside `<MemoryRouter initialEntries={['/path']}>`
- **Wait for async effects**: use `waitFor(() => expect(...))` after render when the component fires an API call in `useEffect`
- **Query buttons by role**: when a text string appears in multiple elements, use `screen.getByRole('button', { name: '...' })` instead of `getByText`

**Run Biome check and frontend tests before committing:**

```sh
cd frontend && npm run check
cd frontend && npm test
```

## Documentation Maintenance

**When making code changes, always update the following documentation in the same commit:**

| Change type | Documents to update |
|-------------|---------------------|
| Add/remove a game | `README.md` (Description, Run section), `CLAUDE.md` (Commands, Games implemented), `AGENTS.md` (Repository Overview, Games implemented) |
| Add/remove a CLI command (`cmd/cli/main.go`) | `README.md` (Run section), `CLAUDE.md` (Commands), `AGENTS.md` (Common Commands) |
| Add/remove a Web API endpoint | `CLAUDE.md` (Web API in Key patterns), `AGENTS.md` (Web API in Key patterns), `api/openapi.yaml` |
| Change request/response schema of a Web API endpoint | `api/openapi.yaml` |
| Change architecture or layer structure | `README.md` (Architecture), `CLAUDE.md` (Architecture), `AGENTS.md` (Architecture) |
| Change Git workflow or CI/CD | `CLAUDE.md` (Git Workflow), `AGENTS.md` (Git Workflow & CI/CD) |
| Modify anything under `frontend/` | Run `cd frontend && npm run check` and `cd frontend && npm test` and ensure both pass before committing |
| Add/remove frontend source files or change testing approach | Update `CLAUDE.md` (Frontend testing) and `AGENTS.md` (Frontend testing) |
| Change frontend tooling or scripts | `frontend/README.md` (Scripts, Tooling) |

Use commit type `docs` (or include doc changes in the same commit as the code change) following the Conventional Commits format.

## Git Workflow & CI/CD

- **`develop`**: Default branch; target for all PRs. CodeQL analysis runs on push/PR.
- **`master`**: Triggers automatic version bump, git tag, and GitHub Release.
- **PR Summary**: When creating a PR, if there is an associated issue, the PR description must explicitly close the issue (e.g., `Closes #123`).

## Commit Message Format

All commit messages must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**

| Type | Description |
|------|-------------|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation only changes |
| `style` | Changes that do not affect code meaning (formatting, etc.) |
| `refactor` | A code change that neither fixes a bug nor adds a feature |
| `perf` | A code change that improves performance |
| `test` | Adding missing tests or correcting existing tests |
| `chore` | Changes to the build process or auxiliary tools |

**Examples:**

```
feat(entities): add new card type to BlackJack
fix(poker): correct hand ranking for flush detection
docs: update README with web deployment instructions
test(blackjack): add tests for dealer bust scenario
refactor(usecases): simplify interactor dependency injection
```

**Rules:**
- The description must be in lowercase and not end with a period.
- Use the imperative mood in the description (e.g., "add feature" not "added feature").
- Breaking changes must include `BREAKING CHANGE:` in the footer or append `!` after the type/scope.

## Workflow Orchestration

- **Plan Node Default**
  - Enter plan mode for ANY non-trivial task (3+ steps or architectural decisions)
  - If something goes sideways, STOP and re-plan immediately – don't keep pushing
  - Use plan mode for verification steps, not just building
  - Write detailed specs upfront to reduce ambiguity

- **Subagent Strategy**
  - Use subagents liberally to keep main context window clean
  - Offload research, exploration, and parallel analysis to subagents
  - For complex problems, throw more compute at it via subagents
  - One task per subagent for focused execution

- **Self-Improvement Loop**
  - After ANY correction from the user: update `tasks/lessons.md` with the pattern
  - Write rules for yourself that prevent the same mistake
  - Ruthlessly iterate on these lessons until mistake rate drops
  - Review lessons at session start for relevant project

- **Verification Before Done**
  - Never mark a task complete without proving it works
  - Diff behavior between main and your changes when relevant
  - Ask yourself: "Would a staff engineer approve this?"
  - Run tests, check logs, demonstrate correctness

- **Demand Elegance (Balanced)**
  - For non-trivial changes: pause and ask "is there a more elegant way?"
  - If a fix feels hacky: "Knowing everything I know now, implement the elegant solution"
  - Skip this for simple, obvious fixes – don't over-engineer
  - Challenge your own work before presenting it

- **Autonomous Bug Fixing**
  - When given a bug report: just fix it. Don't ask for hand-holding
  - Point at logs, errors, failing tests – then resolve them
  - Zero context switching required from the user
  - Go fix failing CI tests without being told how

## Task Management

1. **Plan First**: Write plan to `tasks/todo.md` with checkable items
2. **Verify Plan**: Check in before starting implementation
3. **Track Progress**: Mark items complete as you go
4. **Explain Changes**: High-level summary at each step
5. **Document Results**: Add review section to `tasks/todo.md`
6. **Capture Lessons**: Update `tasks/lessons.md` after corrections

## Core Principles

- **Simplicity First**: Make every change as simple as possible. Impact minimal code.
- **No Laziness**: Find root causes. No temporary fixes. Senior developer standards.
- **Minimal Impact**: Changes should only touch what's necessary. Avoid introducing bugs.
