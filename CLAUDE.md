# CLAUDE.md

Go implementations of 138 trump card game algorithms (blackjack, poker, hearts, klondike, baccarat, ...). Run `go run ./cmd/trumpcards games --short` for the canonical list. Clean Architecture with CLI and Web GUI (React + Go REST API).

## Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [Bun](https://bun.sh/) | 1.3.10 |

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

**When making code changes, always update the following documentation in the same commit:**

| Change type | Documents to update |
|-------------|---------------------|
| Add/remove a game | [`README.md`](README.md) (Description, Run section), [`CLAUDE.md`](CLAUDE.md) (available games list), [`docs/games.md`](docs/games.md), Cloudflare Worker WASM registration (see below) |
| Add/remove a CLI command (`cmd/trumpcards/main.go`) | [`README.md`](README.md) (Run section), [`CLAUDE.md`](CLAUDE.md) (available games list) |
| Add/remove a Web API endpoint | [`docs/architecture.md`](docs/architecture.md) (Web API in Key patterns), [`api/openapi.yaml`](api/openapi.yaml) |
| Change request/response schema of a Web API endpoint | [`api/openapi.yaml`](api/openapi.yaml) |
| Change architecture or layer structure | [`README.md`](README.md) (Architecture), [`CLAUDE.md`](CLAUDE.md) (Architecture), [`docs/architecture.md`](docs/architecture.md) |
| Change Git workflow or CI/CD | [`CLAUDE.md`](CLAUDE.md) (Git Workflow) |
| Modify anything under `frontend/` | Run `cd frontend && bun run build`, `cd frontend && bun run check`, and `cd frontend && bun run test` and ensure all three pass before committing |
| Add/remove frontend source files or change testing approach | Update Testing section in [`frontend/CLAUDE.md`](frontend/CLAUDE.md) |
| Change frontend tooling or scripts | [`frontend/README.md`](frontend/README.md) (Scripts, Tooling) |
| Change game rules or game flow logic | `docs/manual/cui/<game>.md` and `docs/manual/web/<game>.md` for the affected game (follow `docs/manual/cui_template.md` / `docs/manual/web_template.md` format) |
| Add a new game manual | Copy `docs/manual/cui_template.md` → `docs/manual/cui/<game>.md`, `docs/manual/web_template.md` → `docs/manual/web/<game>.md` and fill in game-specific content. Also import in `frontend/src/constants/manualTexts.ts` and add route mapping |
| Change Go testing policy or mock patterns | Update Testing section in [`CLAUDE.md`](CLAUDE.md) and [`internal/CLAUDE.md`](internal/CLAUDE.md) |
| Make an architectural decision that passes the ADR litmus test (see Workflow section) | Add or update an ADR in [`docs/adr/`](docs/adr/) (written in Japanese) and update the index in [`docs/adr/README.md`](docs/adr/README.md) |
| Add/modify exported Go symbol | Ensure GoDoc comment (`// SymbolName description`) is present |
| Add/modify exported TS symbol | Ensure TSDoc comment (`/** description */`) is present |
| Change backend struct/interface/domain logic | Update corresponding UML diagrams in [`docs/design/backend.md`](docs/design/backend.md) (class, sequence, state machine) |
| Change frontend component/hook/API/type | Update corresponding UML diagrams in [`docs/design/frontend.md`](docs/design/frontend.md) (class, sequence, state machine) |

Use commit type `docs` (or include doc changes in the same commit as the code change) following the Conventional Commits format.

### Intermediate design docs

**Do NOT commit intermediate design documents (e.g., `docs/superpowers/specs/`) to the repository.** These documents are not maintained after implementation and become tech debt. Instead:

- **Design specs and brainstorming output**: Post as a comment on the relevant GitHub issue
- **Architecture Decision Records (ADRs)**: These ARE worth committing to `docs/adr/` — they capture the *why* behind decisions and remain valuable long-term

## Cloudflare Workers (WASM)

Games are deployed to Cloudflare Workers as WASM binaries via TinyGo. Three workers split games by category:

| Worker | Entry point | Games |
|--------|-------------|-------|
| **casino** | `cmd/workers/casino/main.go` | Table & poker games (blackjack, baccarat, poker, holdem, omaha, omahahilo, bigo, bigohilo, shortdeck, pineapple, crazypineapple, irishpoker, indianpoker, videopoker, deuceswild, jokerpoker, threecard, fourcardpoker, caribbeanstud, texasholdembonus, ultimatetexasholdem, mississippistud, sevencardstud, paigow, chinesepoker, letitride, reddog, razz, badugi, deucetoseven, spanish21, casinowar, dragontiger, blackjackswitch, oasispoker, russianpoker, casinoholdem, highcardflush, scopa, yaniv, tressette, tichu, bourre, napoleon, mighty, bridge, skat, belote, tarneeb, doudizhu, sheepshead, doppelkopf, mus, tute, sueca, klaverjas) |
| **classic** | `cmd/workers/classic/main.go` | Trick-taking, matching & fishing (hearts, spades, pitch, twotenjack, callbreak, briscola, oldmaid, doubt, daifugo, bigtwo, sevens, crazyeights, ohhell, speed, gofish, pinochle, pigtail, durak, war, fiftyone, whist, pageone, trash, president, cassino, spiteandmalice, shithead, nertz, slapjack, egyptianratscrew, tonk, sixcardgolf, truco) |
| **solo** | `cmd/workers/solo/main.go` | Solitaire & rummy (klondike, freecell, seahaventowers, cruel, spider, spiderette, pyramid, tripeaks, memory, ginrummy, canasta, cribbage, golf, clocksolitaire, fortythieves, canfield, yukon, russiansolitaire, scorpion, wasp, accordion, pokersquares, montecarlo, contractrummy, calculation, bakersdozen, beleagueredcastle, sevenbridge, crescent, gaps, rummy500, eightoff, penguin, acesup, barbu, macau, thirtyone, tienlen, osmosis, fivehundred, schnapsen, burraco, gongzhu, bristol, bidwhist, easthaven, bakersgame, euchre, piquet) |

The worker entry points (`cmd/workers/{casino,classic,solo}/main.go`) are thin shells that blank-import the matching `internal/infrastructure/games/<category>` sub-package and call `games.RegisterCategory(mux, games.Category…)`. The registry itself (`internal/infrastructure/games/registry.go`) stores `{Name, Category, Description}` for each game (the description SSoT — `ui.gameRegistry` reads it from here); the Web-server factories live in `games_server.go` (excluded from WASM via build tags) and the Worker bindings live in per-category sub-packages — this split is what keeps each Cloudflare Worker binary under the 1 MB gzipped free-tier limit by letting TinyGo dead-code-eliminate the games from the other two categories.

**When adding/modifying a game, always update:**
1. `internal/infrastructure/games/registry.go` — `{Name, Category, Description}` entry (selects the worker; Description is the CLI display title)
2. `internal/infrastructure/games/games_server.go` — `BindWebControllerFor("<name>", …)` for the HTTP server factory
3. `internal/infrastructure/games/{casino,classic,solo}/<category>.go` — `games.RegisterKVGame("<name>", games.Category…, …)` for the KV-backed worker route (must match the `Category`)
4. `frontend/src/api/gameApi.ts` `workerUrl` — must match the `Category`

Build: `make build-worker-{solo,casino,classic}` or `make build-workers` (requires TinyGo).

## New Game Addition Checklist

See [`docs/new-game-checklist.md`](docs/new-game-checklist.md) for the full checklist (backend, frontend, docs, verification).

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
