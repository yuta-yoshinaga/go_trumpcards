# Documentation Maintenance

**When making code changes, always update the following documentation in the same commit:**

| Change type | Documents to update |
|-------------|---------------------|
| Add/remove a game | [`README.md`](../README.md) (Description, Run section), [`CLAUDE.md`](../CLAUDE.md) (available games list), [`docs/games.md`](games.md), Cloudflare Worker WASM registration (see [`docs/cloudflare-workers.md`](cloudflare-workers.md)) |
| Add/remove a CLI command (`cmd/trumpcards/main.go`) | [`README.md`](../README.md) (Run section), [`CLAUDE.md`](../CLAUDE.md) (available games list) |
| Add/remove a Web API endpoint | [`docs/architecture.md`](architecture.md) (Web API in Key patterns), [`api/openapi.yaml`](../api/openapi.yaml) |
| Change request/response schema of a Web API endpoint | [`api/openapi.yaml`](../api/openapi.yaml) |
| Change architecture or layer structure | [`README.md`](../README.md) (Architecture), [`CLAUDE.md`](../CLAUDE.md) (Architecture), [`docs/architecture.md`](architecture.md) |
| Change Git workflow or CI/CD | [`CLAUDE.md`](../CLAUDE.md) (Git Workflow) |
| Modify anything under `frontend/` | Run `cd frontend && bun run build`, `cd frontend && bun run check`, and `cd frontend && bun run test` and ensure all three pass before committing |
| Add/remove frontend source files or change testing approach | Update Testing section in [`frontend/CLAUDE.md`](../frontend/CLAUDE.md) |
| Change frontend tooling or scripts | [`frontend/README.md`](../frontend/README.md) (Scripts, Tooling) |
| Change game rules or game flow logic | `docs/manual/cui/<game>.md` and `docs/manual/web/<game>.md` for the affected game (follow `docs/manual/cui_template.md` / `docs/manual/web_template.md` format) |
| Add a new game manual | Copy `docs/manual/cui_template.md` → `docs/manual/cui/<game>.md`, `docs/manual/web_template.md` → `docs/manual/web/<game>.md` and fill in game-specific content. Also import in `frontend/src/constants/manualTexts.ts` and add route mapping |
| Change Go testing policy or mock patterns | Update Testing section in [`CLAUDE.md`](../CLAUDE.md) and [`internal/CLAUDE.md`](../internal/CLAUDE.md) |
| Make an architectural decision that passes the ADR litmus test (see Workflow section) | Add or update an ADR in [`docs/adr/`](adr/) (written in Japanese) and update the index in [`docs/adr/README.md`](adr/README.md) |
| Add/modify exported Go symbol | Ensure GoDoc comment (`// SymbolName description`) is present |
| Add/modify exported TS symbol | Ensure TSDoc comment (`/** description */`) is present |
| Change backend struct/interface/domain logic | Update corresponding UML diagrams in [`docs/design/backend.md`](design/backend.md) (class, sequence, state machine) |
| Change frontend component/hook/API/type | Update corresponding UML diagrams in [`docs/design/frontend.md`](design/frontend.md) (class, sequence, state machine) |

Use commit type `docs` (or include doc changes in the same commit as the code change) following the Conventional Commits format.

## Intermediate design docs

**Do NOT commit intermediate design documents (e.g., `docs/superpowers/specs/`) to the repository.** These documents are not maintained after implementation and become tech debt. Instead:

- **Design specs and brainstorming output**: Post as a comment on the relevant GitHub issue
- **Architecture Decision Records (ADRs)**: These ARE worth committing to `docs/adr/` — they capture the *why* behind decisions and remain valuable long-term
