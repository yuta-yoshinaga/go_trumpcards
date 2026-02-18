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
