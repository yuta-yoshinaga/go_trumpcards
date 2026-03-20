# Architecture

The project implements Clean Architecture with strict layer dependency rules (outer depends on inner, never the reverse). The directory layout follows `golang-standards/project-layout` (`cmd/` + `internal/`):

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
    api/                       # API client functions (fetch wrappers for game endpoints)
    components/                # Shared React components (NavBar, CardImage, CardBack)
    hooks/                     # Custom React hooks (useGameApi, backed by TanStack React Query)
    i18n/                      # i18n config and translation files (ja/en)
    pages/                     # Game page components (BlackJackPage, PokerPage, OldMaidPage)
    providers/                 # React context providers (QueryProvider for TanStack React Query)
    types/                     # TypeScript type definitions for card/game data
  e2e/                         # Playwright E2E test specs
public/                        # Built frontend assets served by Go web server
  assets/                      # Vite-compiled JS/CSS bundles
  images/                      # Card images (PNG)
  css/                         # Bootstrap CSS
```

**Data flow:** `infrastructure` -> `adapter` -> `usecase` -> `domain`

## Key patterns

- **Presenter pattern**: `internal/usecase/presenter/` defines output interfaces (e.g., `BlackJackPresenter`). `internal/adapter/presenter/` provides concrete implementations (CUI vs Web). Presenters are injected into interactors.
- **Mock presenters**: `*_mock.go` files in `internal/usecase/presenter/` are used in tests to avoid I/O.
- **Web API**: Fifteen endpoints -- `POST /blackjack/exec` (BlackJack), `POST /poker/exec` (Poker), `POST /oldmaid/exec` (Old Maid), `POST /daifugo/exec` (Daifugo), `POST /sevens/exec` (Sevens), `POST /doubt/exec` (Doubt), `POST /holdem/exec` (Texas Hold'em), `POST /omaha/exec` (Omaha Hold'em), `POST /hearts/exec` (Hearts), `POST /memory/exec` (Memory), `POST /klondike/exec` (Klondike), `POST /freecell/exec` (FreeCell), `POST /baccarat/exec` (Baccarat), `POST /spades/exec` (Spades), and `POST /crazyeights/exec` (Crazy Eights) -- accept JSON with a `Cmd` field and game state.
