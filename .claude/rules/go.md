---
globs: ["**/*.go"]
---

# Go File Editing Rules

## Format

- **Always run `goimports -w <file>` before committing** (do not use `gofmt`)

## Lint

- **Run `golangci-lint run ./...` before committing and ensure no warnings or errors**

## Testing

**Unit tests are mandatory.** Include them in the same commit as the implementation.

### TDD Cycle (Red → Green → Refactor)

Always follow this cycle before implementing:

1. **Red** — Write a failing test first. Create a test that captures the expected behavior before writing implementation code:
   ```sh
   go test -tags test ./path/to/package -run TestNewFeature  # Fails (Red)
   ```
2. **Green** — Write the minimum code to pass the test. Do not add extra functionality:
   ```sh
   go test -tags test ./path/to/package -run TestNewFeature  # Passes (Green)
   ```
3. **Refactor** — Clean up code while keeping tests green. Improve naming, structure, and remove duplication:
   ```sh
   go test -tags test ./...  # All tests pass (after Refactor)
   ```

Run tests: `go test -tags test ./...`

### Coverage Standard

- `cmd/` and `internal/infrastructure/` are excluded from coverage
- All other packages under `internal/` require **branch coverage (C1) of 80% or higher**
- Focus on business logic and critical paths

### Test Locations (by layer)

| Layer | Test file |
|-------|-----------|
| Domain | `internal/domain/*_test.go` |
| Use cases | `internal/usecase/*Interactor_test.go` |
| Presenters | `internal/adapter/presenter/*_test.go` |
| Controllers | `internal/adapter/controller/*_test.go` |

### Mock Pattern

- **Presenter mocks**: `internal/usecase/presenter/*_mock.go` — implemented with `testify/mock`
- **Interactor mocks**: `internal/adapter/controller/usecase/*_mock.go` — implemented with `testify/mock`
- Use existing `BlackJack*_mock.go` as reference pattern

### Writing Deterministic Tests

Do not depend on shuffle order:

- Set up hands manually with `AddCard`; do not depend on order after `Reset`/`Shuffle`
- Give dealer/CPU scores that prevent automatic draws (e.g., BlackJack dealer >= 17)
- Use retry loops up to 1000 iterations to cover both branches of random decisions

## Architecture

Clean Architecture: `infrastructure` → `adapter` → `usecase` → `domain`

- Domain interfaces: `internal/domain/interfaces/`
- Never reverse the dependency direction (outer layers depend on inner layers)

## Dead Code

- Always remove dead code encountered when modifying code
- Detection tool: `golang.org/x/tools/cmd/deadcode`
- Verify manually before deleting (beware of false positives such as reflection-based calls)
