# CLAUDE.md

Go implementations of 240 trump card game algorithms (blackjack, poker, hearts, klondike, baccarat, ...). Run `go run ./cmd/trumpcards games --short` for the canonical list. Clean Architecture with CLI and Web GUI (React + Go REST API).

## Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [Bun](https://bun.sh/) | 1.3.10 |
| [jq](https://jqlang.github.io/jq/) | any (required by the `.claude/settings.json` commit-gate hooks — they silently no-op without it) |

## Package Manager Rule

**Always use `bun` instead of `npm`/`node`, and `bunx` instead of `npx`.** This project uses Bun as the sole JavaScript runtime, package manager, and script runner. Never invoke `node ./node_modules/...` directly — use `bun` or `bunx` instead.

## Commands

```sh
# Run games
go run ./cmd/trumpcards                    # Interactive mode (switch games with 'switch <game>')
go run ./cmd/trumpcards --start poker      # Interactive mode starting at poker (#1604)
go run ./cmd/trumpcards --lang en          # Interactive mode in English
go run ./cmd/trumpcards <game>             # Run a specific game (e.g., blackjack, poker, holdem)
go run ./cmd/trumpcards --lang en <game>   # Run in English
# Run `go run ./cmd/trumpcards games --short` for the canonical list of game names
# (the SSoT lives in `internal/infrastructure/games/registry.go`).
go run ./cmd/trumpcards games      # List all available games
go run ./cmd/trumpcards games --short  # List game names only (for scripting)
go run ./cmd/trumpcards update     # Self-update to the latest version
go run ./cmd/trumpcards version    # Show version information (equivalent to --version)
go run ./cmd/trumpcards version --short  # Print version number only (machine-readable)
go run ./cmd/trumpcards help       # Show top-level help
go run ./cmd/trumpcards help blackjack  # Show help text for a specific game
go run ./cmd/trumpcards web        # Start REST API + web GUI server (via CLI)
go run ./cmd/trumpcards web --port 3000  # Start web server on custom port
go run ./cmd/trumpcards web --host 127.0.0.1  # Bind to localhost only
go run ./cmd/trumpcards web --open         # Start server and open browser (#1607)
go run ./cmd/trumpcards completion bash  # Generate shell completion script (bash/zsh/fish)
go run ./cmd/server                # Start REST API + web GUI server (direct)

# Test
go test -tags test ./...                                              # Run all Go tests
go test -tags test -coverprofile=coverage.out -covermode=set ./...    # Coverage report

# Format
goimports -w ./...           # Format and organize imports (use goimports, not gofmt)

# Lint
golangci-lint run ./...      # Run Go linter (must pass before commit)

# Frontend
cd frontend && bun install   # Install dependencies
cd frontend && bun run build # Build React app
cd frontend && bun run check # Biome lint + format check
cd frontend && bun run typecheck # TypeScript 7 type check (never bare `tsc` -- that is 5.9; see frontend/CLAUDE.md)
cd frontend && bun run test  # Run Vitest unit tests
cd frontend && bun run e2e   # Run Playwright E2E tests
cd frontend && bun run docs:generate  # Generate TypeDoc documentation

# Docker
docker build -t go_trumpcards .
docker run --rm -d -p 8080:8080 go_trumpcards
```

## Architecture

Clean Architecture: `infrastructure` -> `adapter` -> `usecase` -> `domain`. See [`docs/architecture.md`](docs/architecture.md) for full details.

## Testing

**Unit tests are mandatory. Every implementation must ship with tests in the same commit.**

**All code changes must follow the TDD cycle (Red-Green-Refactor).** See [`internal/CLAUDE.md`](internal/CLAUDE.md) and [`frontend/CLAUDE.md`](frontend/CLAUDE.md) for layer-specific TDD details.

### Coverage standard

**Branch coverage (C1) must be 80% or higher** for all packages except `cmd/` and `internal/infrastructure/` (Go) and for `frontend/src/{api,components,pages,utils}` (TypeScript). Focus testing effort on business logic and critical paths rather than exhaustively covering every conditional branch.

### Self-review checklist

Before marking any task complete:

1. All Go tests pass: `go test -tags test ./...`
2. Go lint passes: `golangci-lint run ./...`
3. Go files formatted: `goimports -w` on modified files
4. Frontend checks pass (if applicable): `cd frontend && bun run build && bun run check && bun run test`
5. Branch coverage is 80%+ for modified packages
6. GoDoc/TSDoc comments present on all new/modified exported symbols

### Detailed rules by layer

- [`internal/CLAUDE.md`](internal/CLAUDE.md) -- Go backend: test locations, mock pattern, deterministic test techniques, lint
- [`frontend/CLAUDE.md`](frontend/CLAUDE.md) -- Frontend: test locations, mock pattern, E2E testing, i18n

## Documentation Maintenance

When making code changes, update the relevant documentation **in the same commit**. The full change-type → docs-to-update mapping (games, CLI commands, Web API, ADRs, UML diagrams, GoDoc/TSDoc, etc.) lives in [`docs/documentation-maintenance.md`](docs/documentation-maintenance.md). Use commit type `docs` (or fold doc updates into the code commit). Never commit intermediate design docs — post them to the GitHub issue instead (ADRs are the exception; those belong in `docs/adr/`).

## Cloudflare Workers (WASM)

Games ship to six Cloudflare Workers (`casino`, `classic`, `solo`, `extra`, `extra2`, `extra3`) as TinyGo WASM binaries, split purely to keep each binary under the 1 MB gzipped free-tier limit ([ADR-0032](docs/adr/0032-fourth-worker-capacity.md) added the fourth, [ADR-0036](docs/adr/0036-fifth-sixth-worker-capacity.md) the fifth and sixth). The `Category` in the registry is a binary-size bucket, **not** a user-facing taxonomy, and there is no overflow bucket: **put a new game in whichever worker currently has the most headroom, measured rather than assumed.** `.claude/skills/rebucket-game/scripts/measure.sh` reports every worker's gzip size, and a recent CI run's `tinygo-build` logs give the same figures without a local TinyGo. Adding/modifying a game touches 5 registration points (registry, `games_server.go`, the category sub-package, the `gameRegistry` CLI wiring in `internal/infrastructure/ui/GameManager.go`, and `gameApi.ts`) plus the per-category count assertions in `registry_test.go`. Full per-worker game list, the build-tag split rationale, and build commands: [`docs/cloudflare-workers.md`](docs/cloudflare-workers.md).

## New Game Addition Checklist

See [`docs/new-game-checklist.md`](docs/new-game-checklist.md) for the full checklist (backend, frontend, docs, verification).

## Git Workflow

- **`develop`**: Default branch; target for all PRs. `golangci-lint` and the CI suite run on push/PR; CodeQL runs on push to `develop` (i.e. after the merge) and weekly — see [ADR-0034](docs/adr/0034-codeql-post-merge.md).
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
- ADR format: Status, Date, Context, Decision, Consequences (see [`docs/adr/README.md`](docs/adr/README.md))

#### ADR Litmus Test

Create an ADR only when **all three** of the following are true:

1. **Were alternatives seriously considered?** — If there was only one option, there is no "decision" to record
2. **Would reversing this decision require changes across multiple files/layers?** — If the impact is small, a commit message is sufficient
3. **Would a new team member 6 months from now ask "why is it done this way?"** — If the answer is obvious from the code, no record needed

Where to document design decisions that don't warrant an ADR:

| Case | Where to write |
|------|---------------|
| Design decisions for a new game | GitHub Issue comment or PR description |
| Reason for refactoring | Commit message body |
| UI design direction | GitHub Issue or PR description |

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

## Design System

Always read [`DESIGN.md`](DESIGN.md) before making any visual or UI decisions. All font choices, colors, spacing, and aesthetic direction are defined there. Do not deviate without explicit user approval. In QA mode, flag any code that doesn't match DESIGN.md.

## Detailed Context

| Topic | File |
|-------|------|
| Architecture & key patterns | [`docs/architecture.md`](docs/architecture.md) |
| Architecture Decision Records | [`docs/adr/`](docs/adr/) |
| Documentation maintenance map (change type → docs to update) | [`docs/documentation-maintenance.md`](docs/documentation-maintenance.md) |
| Cloudflare Workers (per-worker game list, build) | [`docs/cloudflare-workers.md`](docs/cloudflare-workers.md) |
| Game descriptions & entities | [`docs/games.md`](docs/games.md) |
| Backend UML design (class, sequence, state machine) | [`docs/design/backend.md`](docs/design/backend.md) |
| Frontend UML design (class, sequence, state machine) | [`docs/design/frontend.md`](docs/design/frontend.md) |
| Design system (fonts, colors, spacing, motion) | [`DESIGN.md`](DESIGN.md) |
| Go backend rules | [`internal/CLAUDE.md`](internal/CLAUDE.md) |
| Frontend rules | [`frontend/CLAUDE.md`](frontend/CLAUDE.md) |

## Skill routing

When the user's request matches an available skill, ALWAYS invoke it using the Skill
tool as your FIRST action. Do NOT answer directly, do NOT use other tools first.
The skill has specialized workflows that produce better results than ad-hoc answers.

Key routing rules:
- Product ideas, "is this worth building", brainstorming → invoke office-hours
- Bugs, errors, "why is this broken", 500 errors → invoke investigate
- Ship, deploy, push, create PR → invoke ship
- QA, test the site, find bugs → invoke qa
- Code review, check my diff → invoke review
- Update docs after shipping → invoke document-release
- Weekly retro → invoke retro
- Design system, brand → invoke design-consultation
- Visual audit, design polish → invoke design-review
- Architecture review → invoke plan-eng-review
- Save progress, checkpoint, resume → invoke checkpoint
- Code quality, health check → invoke health
- Per-game improvement proposals → GitHub issues ("各ゲームの改善提案", "全ゲームのissueを作って") → invoke game-improve
- New-game candidates → GitHub issues ("追加した方が良いゲームを提案", "新規ゲーム候補をissueに") → invoke propose-games
- Implement a single GitHub issue end-to-end ("issueに着手して", "#NNNN を対応して", "implement issue #N") → invoke improve-issue (explicit `/improve-issue <#>`)
- Clear a whole batch of improvement issues, lowest-effort first ("issueバッチを片付けて", "#NNNN〜#MMMM を全部対応", "残りの改善issueを全部やって") → invoke improve-batch (explicit `/improve-batch <range>`)
