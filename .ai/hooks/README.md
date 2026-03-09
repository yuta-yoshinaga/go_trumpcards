# Agent Guardrail Hooks

These are the verification commands that agents must run before committing changes.

## Go source changes

```sh
goimports -w <modified-files>       # Format and organize imports
go test -tags test ./...            # Run all Go tests
```

## Frontend changes

```sh
cd frontend && npm run build        # Build React app
cd frontend && npm run check        # Biome lint + format check
cd frontend && npm test             # Run Vitest unit tests
```

## All changes

Before any commit, ensure:
1. `goimports -w` has been run on all modified Go files
2. `go test -tags test ./...` passes
3. If frontend files were modified: `cd frontend && npm run build && npm run check && npm test` all pass
