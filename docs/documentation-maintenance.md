# Documentation Maintenance

**When making code changes, always update the following documentation in the same commit:**

| Change type | Documents to update |
|-------------|---------------------|
| Add/remove a game | `README.md` (Description, Run section), `CLAUDE.md` (Commands), `docs/games.md` |
| Add/remove a CLI command (`cmd/cli/main.go`) | `README.md` (Run section), `CLAUDE.md` (Commands) |
| Add/remove a Web API endpoint | `docs/architecture.md` (Web API in Key patterns), `api/openapi.yaml` |
| Change request/response schema of a Web API endpoint | `api/openapi.yaml` |
| Change architecture or layer structure | `README.md` (Architecture), `CLAUDE.md` (Architecture), `docs/architecture.md` |
| Change Git workflow or CI/CD | `CLAUDE.md` (Git Workflow) |
| Modify anything under `frontend/` | Run `cd frontend && npm run build`, `cd frontend && npm run check`, and `cd frontend && npm test` and ensure all three pass before committing |
| Add/remove frontend source files or change testing approach | Update `docs/testing.md` (Frontend testing) and `frontend/CLAUDE.md` |
| Change frontend tooling or scripts | `frontend/README.md` (Scripts, Tooling) |
| Change game rules or game flow logic | `docs/manual/cui/<game>.md` and `docs/manual/web/<game>.md` for the affected game |
| Change Go testing policy or mock patterns | Update `docs/testing.md` and `internal/CLAUDE.md` |

Use commit type `docs` (or include doc changes in the same commit as the code change) following the Conventional Commits format.
