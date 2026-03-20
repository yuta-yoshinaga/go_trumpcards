# CLAUDE.md

Go trump card game algorithms -- BlackJack, Poker, Old Maid, Daifugo, Sevens, Doubt, Texas Hold'em, Hearts, Memory, Klondike, FreeCell, Baccarat. Clean Architecture with CLI and Web GUI (React + Go REST API).

## Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [Bun](https://bun.sh/) | 1.3.10 |

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
go run ./cmd/trumpcards update     # Self-update to the latest version
go run ./cmd/trumpcards web        # Start REST API + web GUI server (via CLI)
go run ./cmd/server         # Start REST API + web GUI server (direct)

# Test
go test -tags test ./...                                              # Run all Go tests
go test -tags test -coverprofile=coverage.out -covermode=atomic ./... # Coverage report

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

### Detailed rules by layer

- [`internal/CLAUDE.md`](internal/CLAUDE.md) -- Go backend: test locations, mock pattern, deterministic test techniques, lint
- [`frontend/CLAUDE.md`](frontend/CLAUDE.md) -- Frontend: test locations, mock pattern, E2E testing, i18n

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
| Modify anything under `frontend/` | Run `cd frontend && bun run build`, `cd frontend && bun run check`, and `cd frontend && bun run test` and ensure all three pass before committing |
| Add/remove frontend source files or change testing approach | Update Testing section in [`CLAUDE.md`](CLAUDE.md) (Frontend testing) and [`frontend/CLAUDE.md`](frontend/CLAUDE.md) |
| Change frontend tooling or scripts | [`frontend/README.md`](frontend/README.md) (Scripts, Tooling) |
| Change game rules or game flow logic | `docs/manual/cui/<game>.md` and `docs/manual/web/<game>.md` for the affected game |
| Change Go testing policy or mock patterns | Update Testing section in [`CLAUDE.md`](CLAUDE.md) and [`internal/CLAUDE.md`](internal/CLAUDE.md) |
| Make an architectural decision (new technology, pattern, or structural change) | Add or update an ADR in [`docs/adr/`](docs/adr/) |

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
| Architecture Decision Records | [`docs/adr/`](docs/adr/) |
| Game descriptions & entities | [`docs/games.md`](docs/games.md) |
| Go backend rules | [`internal/CLAUDE.md`](internal/CLAUDE.md) |
| Frontend rules | [`frontend/CLAUDE.md`](frontend/CLAUDE.md) |
