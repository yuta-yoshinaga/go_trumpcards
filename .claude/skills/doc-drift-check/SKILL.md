---
name: doc-drift-check
description: Check all project documentation for discrepancies with the actual codebase and report findings
# Model-invocable: reports drift (it holds no Write/Edit tool, so even --fix cannot
# rewrite files directly). Useful unprompted after any change that touches docs.
allowed-tools: Read, Grep, Glob, Bash, Agent
argument-hint: "[--fix]"
---

# Documentation Drift Check

Verify that all project documentation matches the actual codebase. Report discrepancies and optionally fix them.

## Scope

Check the following documents against the actual code:

### 1. Game list consistency
- `CLAUDE.md` Commands section — game CLI commands
- `README.md` — game descriptions and run commands
- `docs/games.md` — game entity definitions
- `docs/manual/cui/` and `docs/manual/web/` — per-game manuals
- Actual games: registered in `cmd/trumpcards/main.go` and domain files in `internal/domain/`

### 2. Web API endpoints
- `docs/architecture.md` — endpoint list and count
- `api/openapi.yaml` — endpoint definitions
- Actual endpoints: registered in `internal/infrastructure/web/TrumpCardsWeb.go`

### 3. Frontend
- `frontend/CLAUDE.md` — i18n translation file list
- Actual translation files: `frontend/src/i18n/locales/{ja,en}/*.json`
- Actual pages: `frontend/src/pages/*Page.tsx`

### 4. UML design documents
- `docs/design/backend.md` — class, sequence, state machine diagrams for Go backend
  - Domain structs/interfaces vs actual code in `internal/domain/` and `internal/domain/interfaces/`
  - Game list in diagrams vs actual games
  - Phase constants in state machine diagrams vs actual phase definitions
- `docs/design/frontend.md` — class, sequence, state machine diagrams for React frontend
  - Component/hook/API type names vs actual files in `frontend/src/`
  - Phase enums vs actual definitions in `frontend/src/types/phases.ts`
  - Game route categories vs `frontend/src/constants/gameRoutes.ts`

### 5. ADRs
- `docs/adr/README.md` index table — must list all ADR files in `docs/adr/`
- ADR status consistency

### 6. Version info
- `CLAUDE.md` Requirements table (Go, Node.js, Bun versions)
- `go.mod` Go version
- Actual installed versions

### 7. Auto-memory (MEMORY.md)
- Game list and endpoint count in MEMORY.md vs actual

## Procedure

1. Use parallel Explore agents to check each scope area simultaneously
2. Collect all discrepancies found
3. Report results in a structured format:
   - **Discrepancies found**: file, location, current value, expected value
   - **No discrepancies**: confirmed items

4. If `$ARGUMENTS` contains `--fix`:
   - Fix all discrepancies directly in the files
   - Create a GitHub issue documenting what was fixed
   - Do NOT commit (leave changes unstaged for user review)
5. If `$ARGUMENTS` does NOT contain `--fix`:
   - Create a GitHub issue listing all discrepancies found
   - Do not modify any files
