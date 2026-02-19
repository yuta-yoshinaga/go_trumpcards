# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

**Run the application:**
```sh
go run main.go blackjack  # BlackJack CLI
go run main.go poker    # 5-card Draw Poker CLI
go run main.go oldmaid  # Old Maid CLI
go run main.go web      # Start REST API + web GUI server
```

**Test:**
```sh
go test ./...                          # Run all tests
go test ./entities/...                 # Run tests in a specific package
go test ./entities/ -run TestBlackJack # Run a single test by name
```

**Dependencies:**
```sh
go mod tidy
```

**Frontend (React):**
```sh
cd frontend
npm install        # Install Node.js dependencies
npm run build      # Build React app to public/ (run before starting the web server)
npm run dev        # Start Vite dev server (proxies API to localhost:80)
```

> **Important:** The built assets in `public/assets/` and `public/index.html` are committed to the
> repository so the Go web server can serve them without a separate build step on the server.
> **Whenever you modify anything under `frontend/`, you must run `npm run build` and include the
> updated `public/assets/` and `public/index.html` in the same commit.**

## Architecture

The project implements Clean Architecture with strict layer dependency rules (outer depends on inner, never the reverse):

```
entities/                      # Core business logic (innermost)
usecases/                      # Application business rules (interactors)
  presenters/                  # Presenter interfaces defined here
interface_adapters/            # Convert data between layers
  controllers/                 # Route commands to use cases
  presenters/                  # Implement presenter interfaces for CUI and Web
frameworks_drivers/            # Outermost layer
  ui/                          # CLI runner
  web/                         # REST API server (go-json-rest)
frontend/                      # React frontend source (Vite + React + TypeScript)
  src/
    api/                       # API client functions (fetch wrappers for game endpoints)
    components/                # Shared React components (NavBar, CardImage, CardBack)
    pages/                     # Game page components (BlackJackPage, PokerPage, OldMaidPage)
    types/                     # TypeScript type definitions for card/game data
public/                        # Built frontend assets served by Go web server
  assets/                      # Vite-compiled JS/CSS bundles
  images/                      # Card images (PNG)
  css/                         # Bootstrap CSS
```

**Data flow:** `frameworks_drivers` → `interface_adapters` → `usecases` → `entities`

### Key patterns

- **Presenter pattern**: `usecases/presenters/` defines output interfaces (e.g., `BlackJackPresenter`). `interface_adapters/presenters/` provides concrete implementations (CUI vs Web). Presenters are injected into interactors.
- **Mock presenters**: `*_mock.go` files in `interface_adapters/presenters/` and `usecases/presenters/` are used in tests to avoid I/O.
- **Web API**: Three endpoints — `POST /blackjac/exec` (BlackJack), `POST /poker/exec` (Poker), and `POST /oldmaid/exec` (Old Maid) — accept JSON with a `Cmd` field and game state.

### Games implemented

- **BlackJack**: Entities in `entities/BlackJack.go`, `entities/BlackJackPlayer.go`; interactor in `usecases/BlackJackInteractor.go`
- **Poker (5-card Draw)**: Entities in `entities/Poker.go`, `entities/PokerPlayer.go`; interactor in `usecases/PokerInteractor.go`
- **Old Maid (Babanuki)**: Entities in `entities/OldMaid.go`, `entities/OldMaidPlayer.go`; interactor in `usecases/OldMaidInteractor.go`

## Testing Policy

**Unit tests are mandatory. Every implementation must ship with tests in the same commit.**

### Coverage requirements

When adding or modifying any game logic, provide tests for all four layers:

| Layer | Location | What to test |
|-------|----------|--------------|
| Entities | `entities/*_test.go` | All public methods, edge cases, boundary values |
| Use cases | `usecases/*Interactor_test.go` | Each interactor method via a mock presenter |
| Presenters | `interface_adapters/presenters/*_test.go` | CUI text output and Web JSON output for every game phase |
| Controllers | `interface_adapters/controllers/*_test.go` | Every supported command including unknown/empty input |

### Mock pattern

- **Presenter mocks**: `usecases/presenters/*_mock.go` — implement the presenter interface using `testify/mock`
- **Interactor mocks**: `interface_adapters/controllers/usecases/*_mock.go` — implement the interactor interface using `testify/mock`
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

## Documentation Maintenance

**When making code changes, always update the following documentation in the same commit:**

| Change type | Documents to update |
|-------------|---------------------|
| Add/remove a game | `README.md` (Description, Run section), `CLAUDE.md` (Commands, Games implemented), `AGENTS.md` (Repository Overview, Games implemented) |
| Add/remove a CLI command (`main.go`) | `README.md` (Run section), `CLAUDE.md` (Commands), `AGENTS.md` (Common Commands) |
| Add/remove a Web API endpoint | `CLAUDE.md` (Web API in Key patterns), `AGENTS.md` (Web API in Key patterns) |
| Change architecture or layer structure | `README.md` (Architecture), `CLAUDE.md` (Architecture), `AGENTS.md` (Architecture) |
| Change Git workflow or CI/CD | `CLAUDE.md` (Git Workflow), `AGENTS.md` (Git Workflow & CI/CD) |
| Modify anything under `frontend/` | Run `npm run build` and commit updated `public/assets/` and `public/index.html` in the same commit |

Use commit type `docs` (or include doc changes in the same commit as the code change) following the Conventional Commits format.

## Git Workflow

- **`develop`**: Default branch; target for all PRs. CodeQL analysis runs on push/PR.
- **`master`**: Triggers automatic version bump, git tag, and GitHub Release.

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
