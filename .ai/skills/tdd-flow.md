# TDD Flow (Red-Green-Refactor)

**All code changes must follow the TDD cycle:**

## 1. Red -- Write a failing test first

Before writing any production code, create or modify a test file (`*_test.go`) that captures the expected behavior. Run the test and confirm it fails:

```sh
go test -tags test ./path/to/package -run TestNewFeature  # Confirm the test fails (Red)
```

Write tests in the corresponding location for each Clean Architecture layer:

| Layer | Test location |
|-------|--------------|
| Domain | `internal/domain/*_test.go` |
| Use cases | `internal/usecase/*Interactor_test.go` |
| Presenters | `internal/adapter/presenter/*_test.go` |
| Controllers | `internal/adapter/controller/*_test.go` |

## 2. Green -- Write the minimum code to pass

Implement only the code necessary to make the failing test pass. Do not add extra functionality beyond what the test requires:

```sh
go test -tags test ./path/to/package -run TestNewFeature  # Confirm the test passes (Green)
```

## 3. Refactor -- Clean up while keeping tests green

Improve code quality (naming, structure, duplication removal) without changing behavior. Verify all tests still pass after refactoring:

```sh
go test -tags test ./...  # Confirm all tests still pass after refactoring
```

## Key rules

- Never write production code without a corresponding failing test first
- Each Red-Green-Refactor cycle should be small and focused
- Run `go test` at every stage transition to verify the expected outcome
- Apply this cycle at every layer of the Clean Architecture (Domain, Use cases, Presenters, Controllers)

## Self-review checklist

Before marking any task complete:

1. All tests pass: `go test -tags test ./...`
2. Go files formatted: `goimports -w` on modified files
3. Frontend checks pass (if applicable): `cd frontend && npm run build && npm run check && npm test`
4. Branch coverage is 100% for modified packages (excluding `cmd/` and `internal/infrastructure/`)
