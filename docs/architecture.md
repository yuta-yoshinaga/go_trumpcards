# Architecture

The project implements Clean Architecture with strict layer dependency rules (outer depends on inner, never the reverse). The directory layout follows `golang-standards/project-layout` (`cmd/` + `internal/`):

```
cmd/
  trumpcards/                    # CLI entrypoint (all games + web server)
  server/                        # Web server dedicated entrypoint
  workers/                       # Cloudflare Workers WASM entrypoints
    casino/main.go               # Table & poker games
    classic/main.go              # Trick-taking & matching games
    solo/main.go                 # Solitaire & rummy games
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
      cli/                     # CLI mode components (CliTerminal, CliToggle)
    hooks/                     # Custom React hooks (useGameApi, backed by TanStack React Query)
    i18n/                      # i18n config and translation files (ja/en)
    pages/                     # Game page components (BlackJackPage, PokerPage, OldMaidPage)
    providers/                 # React context providers (QueryProvider for TanStack React Query)
    types/                     # TypeScript type definitions for card/game data
    utils/cli/                 # CLI mode utilities
      commands/                # Per-game command parsers (text input → API args)
      formatters/              # Per-game response formatters (JSON → terminal text)
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
- **Web API**: Fifty-six endpoints -- `POST /blackjack/exec` (BlackJack), `POST /poker/exec` (Poker), `POST /oldmaid/exec` (Old Maid), `POST /daifugo/exec` (Daifugo), `POST /sevens/exec` (Sevens), `POST /doubt/exec` (Doubt), `POST /holdem/exec` (Texas Hold'em), `POST /omaha/exec` (Omaha Hold'em), `POST /shortdeck/exec` (Short Deck Hold'em), `POST /pineapple/exec` (Pineapple Poker), `POST /hearts/exec` (Hearts), `POST /memory/exec` (Memory), `POST /klondike/exec` (Klondike), `POST /freecell/exec` (FreeCell), `POST /baccarat/exec` (Baccarat), `POST /spades/exec` (Spades), `POST /twotenjack/exec` (Two Ten Jack), `POST /crazyeights/exec` (Crazy Eights), `POST /ginrummy/exec` (Gin Rummy), `POST /spider/exec` (Spider Solitaire), `POST /napoleon/exec` (Napoleon), `POST /indianpoker/exec` (Indian Poker), `POST /videopoker/exec` (Video Poker), `POST /deuceswild/exec` (Deuces Wild), `POST /jokerpoker/exec` (Joker Poker), `POST /euchre/exec` (Euchre), `POST /pyramid/exec` (Pyramid), `POST /tripeaks/exec` (TriPeaks), `POST /cribbage/exec` (Cribbage), `POST /threecard/exec` (Three Card Poker), `POST /caribbeanstud/exec` (Caribbean Stud Poker), `POST /ohhell/exec` (Oh Hell), `POST /bridge/exec` (Contract Bridge), `POST /speed/exec` (Speed), `POST /gofish/exec` (Go Fish), `POST /canasta/exec` (Canasta), `POST /pinochle/exec` (Pinochle), `POST /golf/exec` (Golf Solitaire), `POST /pigtail/exec` (Pig's Tail), `POST /sevencardstud/exec` (Seven Card Stud), `POST /clocksolitaire/exec` (Clock Solitaire), `POST /durak/exec` (Durak), `POST /fortythieves/exec` (Forty Thieves), `POST /paigow/exec` (Pai Gow Poker), `POST /war/exec` (War), `POST /canfield/exec` (Canfield), `POST /fiftyone/exec` (Fifty-one), `POST /yukon/exec` (Yukon), `POST /scorpion/exec` (Scorpion), `POST /whist/exec` (Whist), `POST /letitride/exec` (Let It Ride), `POST /pokersquares/exec` (Poker Squares), `POST /pageone/exec` (Page One), `POST /reddog/exec` (Red Dog), `POST /razz/exec` (Razz), and `POST /badugi/exec` (Badugi) -- accept JSON with a `Cmd` field and game state.
- **Swagger UI**: Available at `/swagger/` -- serves the OpenAPI spec (`api/openapi.yaml`) via Swagger UI for interactive API documentation and testing. The spec is embedded into the binary with `go:embed`; the Swagger UI frontend is loaded from a CDN.

## Cloudflare Workers (WASM) deployment

In addition to Docker deployment (Render), the project supports edge deployment via Cloudflare Workers. See [ADR-0027](adr/0027-cloudflare-workers-wasm.md) and [ADR-0028](adr/0028-kv-session-persistence.md) for decision rationale.

### Build toolchain

Go source (`cmd/workers/*/main.go`) is compiled to WASM using **TinyGo** (not the standard Go compiler) to meet Cloudflare's 1MB compressed size limit. The pipeline is:

```
Go source (//go:build js && wasm)
  → TinyGo compiler (tinygo build -target wasm)
  → wasm-opt (size optimisation)
  → worker.mjs + wasm_exec.js (JS runtime wrapper, generated by workers-assets-gen)
  → wrangler deploy (Cloudflare Workers)
```

Build commands: `make build-worker-{solo,casino,classic}` or `make build-workers`.

### 3-Worker split

Games are distributed across three Workers to stay under the 1MB gzip size limit per Worker:

| Worker | Category | Entry point | Games |
|--------|----------|-------------|-------|
| **casino** | Table & poker | `cmd/workers/casino/main.go` | blackjack, baccarat, poker, holdem, omaha, shortdeck, pineapple, indianpoker, videopoker, deuceswild, jokerpoker, threecard, caribbeanstud, sevencardstud, paigow, letitride, reddog, razz, badugi |
| **classic** | Trick-taking & matching | `cmd/workers/classic/main.go` | hearts, spades, euchre, napoleon, oldmaid, doubt, daifugo, sevens, crazyeights, ohhell, bridge, speed, gofish, pinochle, pigtail, twotenjack, war, durak, fiftyone, whist |
| **solo** | Solitaire & rummy | `cmd/workers/solo/main.go` | klondike, freecell, spider, pyramid, tripeaks, memory, ginrummy, canasta, cribbage, golf, clocksolitaire, fortythieves, canfield, yukon, scorpion, pokersquares |

The frontend routes requests to the correct Worker via `workerUrl` mapping in `frontend/src/api/gameApi.ts`. When `VITE_WORKER_*_URL` env vars are unset, requests fall back to relative URLs (Docker deployment).

### Session persistence

Workers are stateless, so game sessions are persisted in **Cloudflare KV** (`GAME_SESSIONS` namespace). Each game's domain object implements `MarshalJSON`/`UnmarshalJSON` for serialisation. The `SessionProvider[T]` interface abstracts session storage:

- **Docker**: `MemorySessionProvider` (in-memory `SessionStore[T]`)
- **Workers**: `KVSessionProvider` (Cloudflare KV, TTL=1h, JSON serialised)

### Adding a game to Workers

When adding a new game, register it in the appropriate Worker entry point using `registerKV`:

```go
// cmd/workers/<worker>/main.go
registerKV(mux, "/<game>/exec", "<game>:",
    func() usecase.<Game>InteractorIF { ... },      // factory
    func(data []byte) (usecase.<Game>InteractorIF, error) { ... }, // restore
    func(p controller.SessionProvider[usecase.<Game>InteractorIF], f func() usecase.<Game>InteractorIF) interface { ... } { ... }, // controller
)
```

Also ensure `frontend/src/api/gameApi.ts` `workerUrl` maps the game to the correct `WORKER_*` constant.

### TinyGo constraints

- `go.mod` specifies `go 1.25.8` (TinyGo's latest supported Go version) with `toolchain go1.26.0` for local development
- Mock files require `//go:build test` tag to exclude `testify/mock` from WASM builds
- `net/http` method-prefixed routing (`"POST /path"`) is not supported; Worker entry points use plain `"/path"` patterns
