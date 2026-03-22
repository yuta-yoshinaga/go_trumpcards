# CLAUDE.md

Go trump card game algorithms -- BlackJack, Poker, Old Maid, Daifugo, Sevens, Doubt, Texas Hold'em, Omaha Hold'em, Hearts, Memory, Klondike, FreeCell, Baccarat, Spades, Crazy Eights, Gin Rummy, Napoleon. Clean Architecture with CLI and Web GUI (React + Go REST API).

## Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [Bun](https://bun.sh/) | 1.3.10 |

## Resource Constraints

This development environment runs on WSL2 with **limited RAM (~2 GB) and 4 CPU cores**. Heavy tasks (frontend tests, Go tests, builds, linting) must NOT be launched in parallel — doing so causes SWAP thrashing and dramatically slows overall execution.

### Rules

1. **Kill residual processes before launching heavy tasks.** Previous processes may not have fully exited. Always run `pkill` first to ensure no leftover processes compete for memory:
   ```sh
   pkill -f vitest || true; pkill -f 'bun run' || true; pkill -f 'go test' || true; pkill -f golangci-lint || true
   ```
   Run this before every heavy task invocation (build, test, lint).

2. **Run heavy tasks sequentially, not in parallel.** Chain with `&&`:
   ```sh
   cd frontend && bun run build && bun run check && bun run test
   ```
   Do NOT launch `bun run build`, `bun run check`, and `bun run test` as separate parallel tool calls or background tasks.

3. **Limit Vitest worker threads** when running tests on their own:
   ```sh
   cd frontend && bun run test -- --pool-options.threads.maxThreads=2
   ```

4. **Limit Go test parallelism**:
   ```sh
   go test -tags test -p 2 ./...
   ```

5. **Never launch frontend and Go tasks simultaneously** (e.g., `bun run test` and `go test` at the same time).

6. **Avoid multiple background tasks** (`run_in_background`) for resource-heavy commands. Use background only for lightweight commands (e.g., `git`, `ls`, `gh`).

## Package Manager Rule

**Always use `bun` instead of `npm`/`node`, and `bunx` instead of `npx`.** This project uses Bun as the sole JavaScript runtime, package manager, and script runner. Never invoke `node ./node_modules/...` directly — use `bun` or `bunx` instead, as Node.js consumes significantly more memory.

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
go run ./cmd/trumpcards omaha      # Omaha Hold'em CLI
go run ./cmd/trumpcards hearts     # Hearts CLI
go run ./cmd/trumpcards memory     # Memory (神経衰弱) CLI
go run ./cmd/trumpcards klondike   # Klondike (ソリティア) CLI
go run ./cmd/trumpcards freecell   # FreeCell CLI
go run ./cmd/trumpcards baccarat   # Baccarat (バカラ) CLI
go run ./cmd/trumpcards spades     # Spades (スペード) CLI
go run ./cmd/trumpcards crazyeights   # Crazy Eights CLI
go run ./cmd/trumpcards ginrummy   # Gin Rummy (ジンラミー) CLI
go run ./cmd/trumpcards spider     # Spider Solitaire (スパイダーソリティア) CLI
go run ./cmd/trumpcards napoleon   # Napoleon (ナポレオン) CLI
go run ./cmd/trumpcards update     # Self-update to the latest version
go run ./cmd/trumpcards web        # Start REST API + web GUI server (via CLI)
go run ./cmd/server         # Start REST API + web GUI server (direct)

# Test
go test -tags test -p 2 ./...                                              # Run all Go tests
go test -tags test -p 2 -coverprofile=coverage.out -covermode=atomic ./... # Coverage report

# Format
goimports -w ./...           # Format and organize imports (use goimports, not gofmt)

# Lint
golangci-lint run ./...      # Run Go linter (must pass before commit)

# Frontend
cd frontend && bun install   # Install dependencies
cd frontend && bun run build # Build React app
cd frontend && bun run check # Biome lint + format check
cd frontend && bun run test  # Run Vitest unit tests
cd frontend && bun run e2e   # Run Playwright E2E tests
cd frontend && bun run docs:generate  # Generate TypeDoc documentation

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

1. **Red** -- Write a failing test first. Confirm it fails before writing production code.
2. **Green** -- Write the minimum code to make the test pass.
3. **Refactor** -- Clean up without changing behavior. Verify all tests still pass.

### Coverage standard

**Branch coverage (C1) must be 100%** for all packages except `cmd/` and `internal/infrastructure/` (Go) and for `frontend/src/{api,components,pages,utils}` (TypeScript). Always verify every conditional branch is exercised, not just statement coverage (C0).

### Self-review checklist

Before marking any task complete:

1. All Go tests pass: `go test -tags test ./...`
2. Go lint passes: `golangci-lint run ./...`
3. Go files formatted: `goimports -w` on modified files
4. Frontend checks pass (if applicable): `cd frontend && bun run build && bun run check && bun run test`
5. Branch coverage is 100% for modified packages
6. GoDoc/TSDoc comments present on all new/modified exported symbols

### Detailed rules by layer

- [`internal/CLAUDE.md`](internal/CLAUDE.md) -- Go backend: test locations, mock pattern, deterministic test techniques, lint
- [`frontend/CLAUDE.md`](frontend/CLAUDE.md) -- Frontend: test locations, mock pattern, E2E testing, i18n

## Documentation Maintenance

**When making code changes, always update the following documentation in the same commit:**

| Change type | Documents to update |
|-------------|---------------------|
| Add/remove a game | [`README.md`](README.md) (Description, Run section), [`CLAUDE.md`](CLAUDE.md) (Commands), [`docs/games.md`](docs/games.md) |
| Add/remove a CLI command (`cmd/trumpcards/main.go`) | [`README.md`](README.md) (Run section), [`CLAUDE.md`](CLAUDE.md) (Commands) |
| Add/remove a Web API endpoint | [`docs/architecture.md`](docs/architecture.md) (Web API in Key patterns), [`api/openapi.yaml`](api/openapi.yaml) |
| Change request/response schema of a Web API endpoint | [`api/openapi.yaml`](api/openapi.yaml) |
| Change architecture or layer structure | [`README.md`](README.md) (Architecture), [`CLAUDE.md`](CLAUDE.md) (Architecture), [`docs/architecture.md`](docs/architecture.md) |
| Change Git workflow or CI/CD | [`CLAUDE.md`](CLAUDE.md) (Git Workflow) |
| Modify anything under `frontend/` | Run `cd frontend && bun run build`, `cd frontend && bun run check`, and `cd frontend && bun run test` and ensure all three pass before committing |
| Add/remove frontend source files or change testing approach | Update Testing section in [`CLAUDE.md`](CLAUDE.md) (Frontend testing) and [`frontend/CLAUDE.md`](frontend/CLAUDE.md) |
| Change frontend tooling or scripts | [`frontend/README.md`](frontend/README.md) (Scripts, Tooling) |
| Change game rules or game flow logic | `docs/manual/cui/<game>.md` and `docs/manual/web/<game>.md` for the affected game (follow `docs/manual/cui_template.md` / `docs/manual/web_template.md` format) |
| Add a new game manual | Copy `docs/manual/cui_template.md` → `docs/manual/cui/<game>.md`, `docs/manual/web_template.md` → `docs/manual/web/<game>.md` and fill in game-specific content |
| Change Go testing policy or mock patterns | Update Testing section in [`CLAUDE.md`](CLAUDE.md) and [`internal/CLAUDE.md`](internal/CLAUDE.md) |
| Make an architectural decision (new technology, pattern, or structural change) | Add or update an ADR in [`docs/adr/`](docs/adr/) (日本語で記述) and update the index in [`docs/adr/README.md`](docs/adr/README.md) |
| Add/modify exported Go symbol | Ensure GoDoc comment (`// SymbolName description`) is present |
| Add/modify exported TS symbol | Ensure TSDoc comment (`/** description */`) is present |
| Change backend struct/interface/domain logic | Update corresponding UML diagrams in [`docs/design/backend.md`](docs/design/backend.md) (class, sequence, state machine) |
| Change frontend component/hook/API/type | Update corresponding UML diagrams in [`docs/design/frontend.md`](docs/design/frontend.md) (class, sequence, state machine) |

Use commit type `docs` (or include doc changes in the same commit as the code change) following the Conventional Commits format.

### Intermediate design docs

**Do NOT commit intermediate design documents (e.g., `docs/superpowers/specs/`) to the repository.** These documents are not maintained after implementation and become tech debt. Instead:

- **Design specs and brainstorming output**: Post as a comment on the relevant GitHub issue
- **Architecture Decision Records (ADRs)**: These ARE worth committing to `docs/adr/` — they capture the *why* behind decisions and remain valuable long-term

## New Game Addition Checklist

When adding a new game, follow this checklist to avoid post-feat fix commits. Complete ALL items before creating the PR.

### Backend (Go)

1. **Domain**: Create `internal/domain/<Game>.go`, `<Game>Player.go`, `<Game>Config.go` (if configurable)
2. **Reuse shared helpers**: `deal_helper.go` (dealAllCards), `hand_eval.go` (hand evaluation), `betting.go` (chip/betting), `play_style_helper.go` (CPU styles), `player_helpers.go` (resetPlayers), `hesitation.go` (CPU delay), `memory_manager.go`/`memory_decay.go` (CPU memory AI), `GamePlayer.go` (base player struct), `ChipHolder.go` (chip system), `kicker.go` (kicker comparison)
3. **Interactor**: `internal/usecase/<Game>Interactor.go` with presenter interface in `internal/usecase/presenter/`
4. **Controller**: CUI controller in `internal/adapter/controller/`, Web controller in `internal/adapter/controller/`, reuse `cuiutil` package for input parsing and `ClampIntPtr` for config validation
5. **Presenter**: CUI and Web presenters in `internal/adapter/presenter/`, reuse `buildCuiOutput`, `cuiCardListStr`, `ActionLogOutput` helpers, `WebOutputBase` for common web output fields
6. **Infrastructure**: Register in `cmd/trumpcards/main.go` (CLI) and `internal/infrastructure/web/TrumpCardsWeb.go` (API route)
7. **Run `goimports -w` and `golangci-lint run ./...`** on all new files
8. **100% branch coverage** for all new packages

### Frontend (React)

9. **Page**: `frontend/src/pages/<Game>Page.tsx` with test file, reuse `useGamePageSetup` hook, `usePhaseNames`, `gameReplay`, `useCardDimensions`, `gameExec` API helper
10. **Shared components**: Use `PhaseIndicator`, `SettingsPanel`, `ConfirmDialog`, `ActionLogSection`, `GameFooter`, `GameMessageBox`, `AnimatedCardBack`, `ErrorBoundary`
11. **i18n**: Add `frontend/src/i18n/locales/{ja,en}/<game>.json` translation files
12. **Router**: Add route in `frontend/src/App.tsx` and NavBar entry
13. **Run `bun run build && bun run check && bun run test`**

### Documentation (same commit)

14. **`README.md`**: Add game description and CLI command
15. **`CLAUDE.md`**: Add CLI command to Commands section
16. **`docs/games.md`**: Add game entity description
17. **`docs/architecture.md`**: Update endpoint count and list
18. **`api/openapi.yaml`**: Add endpoint path, tag definition, and request/response schemas in components
19. **`frontend/CLAUDE.md`**: Update i18n translation file list
20. **`docs/manual/cui/<game>.md`** and **`docs/manual/web/<game>.md`**: Add game manuals

### Final verification

21. `go test -tags test ./...` -- all tests pass
22. `golangci-lint run ./...` -- no warnings
23. `cd frontend && bun run build && bun run check && bun run test` -- all pass
24. **E2E test**: Create `frontend/e2e/<game>.spec.ts` with basic game flow test
25. `cd frontend && bun run e2e` -- all E2E tests pass

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

### ADR-Driven Decision Making

- **Before starting any architectural change**, review relevant ADRs in [`docs/adr/`](docs/adr/) to understand past decisions and their rationale
- When a change contradicts or supersedes an existing ADR, update the ADR's status to `Superseded` and create a new ADR documenting the new decision
- When making a new architectural decision (technology choice, structural change, new pattern), create a new ADR before or alongside the implementation
- ADR format: Status, Date, Context, Decision, Consequences (see [`docs/adr/README.md`](docs/adr/README.md))

### Plan Node Default

- Enter plan mode for ANY non-trivial task (3+ steps or architectural decisions)
- If something goes sideways, STOP and re-plan immediately -- don't keep pushing
- Use plan mode for verification steps, not just building
- Write detailed specs upfront to reduce ambiguity

### Core Principles

- **Simplicity First**: Make every change as simple as possible. Impact minimal code.
- **No Laziness**: Find root causes. No temporary fixes. Senior developer standards.
- **Minimal Impact**: Changes should only touch what's necessary. Avoid introducing bugs.
- **Dead Code Cleanup**: When modifying code, always remove any dead code or dead files you encounter. Use `golang.org/x/tools/cmd/deadcode` for Go and `knip` for TypeScript to identify unused code. Verify findings manually before deleting -- static analysis tools can produce false positives (e.g., interface implementations called via reflection, mock methods). Delete confirmed dead code in the same commit as your feature or fix.

## Detailed Context

| Topic | File |
|-------|------|
| Architecture & key patterns | [`docs/architecture.md`](docs/architecture.md) |
| Architecture Decision Records | [`docs/adr/`](docs/adr/) |
| Game descriptions & entities | [`docs/games.md`](docs/games.md) |
| Backend UML design (class, sequence, state machine) | [`docs/design/backend.md`](docs/design/backend.md) |
| Frontend UML design (class, sequence, state machine) | [`docs/design/frontend.md`](docs/design/frontend.md) |
| Go backend rules | [`internal/CLAUDE.md`](internal/CLAUDE.md) |
| Frontend rules | [`frontend/CLAUDE.md`](frontend/CLAUDE.md) |
