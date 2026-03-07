# CLAUDE.md

Go trump card game algorithms -- BlackJack, Poker, Old Maid, Daifugo, Sevens, Doubt, Texas Hold'em. Clean Architecture with CLI and Web GUI (React + Go REST API).

## Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [npm](https://www.npmjs.com/) | 11.x |

## Commands

```sh
# Run games
go run ./cmd/cli blackjack  # BlackJack CLI
go run ./cmd/cli poker      # 5-card Draw Poker CLI
go run ./cmd/cli oldmaid    # Old Maid CLI
go run ./cmd/cli daifugo    # Daifugo CLI
go run ./cmd/cli sevens     # Sevens (7並べ) CLI
go run ./cmd/cli doubt      # Doubt (ダウト) CLI
go run ./cmd/cli holdem     # Texas Hold'em CLI
go run ./cmd/cli web        # Start REST API + web GUI server (via CLI)
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

TDD (Red-Green-Refactor) is mandatory. 100% branch coverage (C1) required. See:
- [`docs/testing.md`](docs/testing.md) -- Full testing policy (Go + frontend + E2E + i18n)
- [`.ai/skills/tdd-flow.md`](.ai/skills/tdd-flow.md) -- TDD cycle workflow
- [`.ai/hooks/README.md`](.ai/hooks/README.md) -- Pre-commit verification commands
- [`internal/CLAUDE.md`](internal/CLAUDE.md) -- Go-specific testing rules
- [`frontend/CLAUDE.md`](frontend/CLAUDE.md) -- Frontend-specific testing rules

## Documentation Maintenance

See [`docs/documentation-maintenance.md`](docs/documentation-maintenance.md) for the update matrix.

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

See [`docs/workflow.md`](docs/workflow.md) for workflow orchestration, task management, and core principles.

## Detailed Context

| Topic | File |
|-------|------|
| Architecture & key patterns | [`docs/architecture.md`](docs/architecture.md) |
| Testing policy (all layers) | [`docs/testing.md`](docs/testing.md) |
| Game descriptions & entities | [`docs/games.md`](docs/games.md) |
| Documentation update matrix | [`docs/documentation-maintenance.md`](docs/documentation-maintenance.md) |
| Workflow & principles | [`docs/workflow.md`](docs/workflow.md) |
| TDD skill | [`.ai/skills/tdd-flow.md`](.ai/skills/tdd-flow.md) |
| Pre-commit hooks | [`.ai/hooks/README.md`](.ai/hooks/README.md) |
| Go backend rules | [`internal/CLAUDE.md`](internal/CLAUDE.md) |
| Frontend rules | [`frontend/CLAUDE.md`](frontend/CLAUDE.md) |
