# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

**Run the application:**
```sh
go run main.go cui     # BlackJack CLI
go run main.go poker   # 5-card Draw Poker CLI
go run main.go web     # Start REST API + web GUI server
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
public/                        # Static web assets (HTML/CSS/JS)
```

**Data flow:** `frameworks_drivers` → `interface_adapters` → `usecases` → `entities`

### Key patterns

- **Presenter pattern**: `usecases/presenters/` defines output interfaces (e.g., `BlackJackPresenter`). `interface_adapters/presenters/` provides concrete implementations (CUI vs Web). Presenters are injected into interactors.
- **Mock presenters**: `*_mock.go` files in `interface_adapters/presenters/` and `usecases/presenters/` are used in tests to avoid I/O.
- **Web API**: Two endpoints — `POST /blackjac/exec` (BlackJack) and `POST /poker/exec` (Poker) — accept JSON with a `Cmd` field and game state.

### Games implemented

- **BlackJack**: Entities in `entities/BlackJack.go`, `entities/BlackJackPlayer.go`; interactor in `usecases/BlackJackInteractor.go`
- **Poker (5-card Draw)**: Entities in `entities/Poker.go`, `entities/PokerPlayer.go`; interactor in `usecases/PokerInteractor.go`

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
