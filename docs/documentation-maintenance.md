# Documentation Maintenance

**When making code changes, always update the following documentation in the same commit:**

| Change type | Documents to update |
|-------------|---------------------|
| Add/remove a game | [`README.md`](../README.md) (Description, Run section), [`CLAUDE.md`](../CLAUDE.md) (Commands), [`docs/games.md`](games.md) |
| Add/remove a CLI command (`cmd/cli/main.go`) | [`README.md`](../README.md) (Run section), [`CLAUDE.md`](../CLAUDE.md) (Commands) |
| Add/remove a Web API endpoint | [`docs/architecture.md`](architecture.md) (Web API in Key patterns), [`api/openapi.yaml`](../api/openapi.yaml) |
| Change request/response schema of a Web API endpoint | [`api/openapi.yaml`](../api/openapi.yaml) |
| Change architecture or layer structure | [`README.md`](../README.md) (Architecture), [`CLAUDE.md`](../CLAUDE.md) (Architecture), [`docs/architecture.md`](architecture.md) |
| Change Git workflow or CI/CD | [`CLAUDE.md`](../CLAUDE.md) (Git Workflow) |
| Modify anything under `frontend/` | Run `cd frontend && npm run build`, `cd frontend && npm run check`, and `cd frontend && npm test` and ensure all three pass before committing |
| Add/remove frontend source files or change testing approach | Update [`docs/testing.md`](testing.md) (Frontend testing) and [`frontend/CLAUDE.md`](../frontend/CLAUDE.md) |
| Change frontend tooling or scripts | [`frontend/README.md`](../frontend/README.md) (Scripts, Tooling) |
| Change game rules or game flow logic | `docs/manual/cui/<game>.md` and `docs/manual/web/<game>.md` for the affected game |
| Change Go testing policy or mock patterns | Update [`docs/testing.md`](testing.md) and [`internal/CLAUDE.md`](../internal/CLAUDE.md) |

Use commit type `docs` (or include doc changes in the same commit as the code change) following the Conventional Commits format.
