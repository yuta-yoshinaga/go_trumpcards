# internal/ -- Go Backend Rules

This directory contains all Go backend code following Clean Architecture.

## Layer structure

| Layer | Path | Depends on |
|-------|------|------------|
| Domain | `domain/` | Nothing (innermost) |
| Use cases | `usecase/` | Domain |
| Adapter | `adapter/` | Use cases, Domain |
| Infrastructure | `infrastructure/` | Adapter, Use cases, Domain |

**Data flow:** `infrastructure` -> `adapter` -> `usecase` -> `domain`

## Testing

**TDD cycle (Red-Green-Refactor):**

1. **Red** -- Write a failing test (`go test -tags test ./internal/domain -run TestName` confirms failure)
2. **Green** -- Write the minimum code to pass the test
3. **Refactor** -- Clean up while keeping all tests green (`go test -tags test ./...`)

**Branch coverage (C1) must be 80% or higher** for all packages in this directory except `infrastructure/` (the top-level `cmd/` directory is also excluded project-wide). Focus testing effort on business logic and critical paths.

### Test locations by layer

| Layer | Test location |
|-------|--------------|
| Domain | `domain/*_test.go` |
| Use cases | `usecase/*Interactor_test.go` |
| Presenters | `adapter/presenter/*_test.go` |
| Controllers | `adapter/controller/*_test.go` |

### Mock pattern

- **Presenter mocks**: `usecase/presenter/*_mock.go` -- implement the presenter interface using `testify/mock`
- **Interactor mocks**: `adapter/controller/usecase/*_mock.go` -- implement the interactor interface using `testify/mock`
- Follow the existing `BlackJack*_mock.go` files as the reference pattern

### Writing deterministic tests

Card games involve shuffling, so tests must not depend on random outcomes:

- **Avoid auto-hit/draw**: Give the dealer/CPU a score that prevents automatic card draws (e.g., BlackJack dealer >= 17, Poker dealer rank >= Two Pair so `dealerExchange` is skipped)
- **Force deterministic draws in OldMaid**: Give the target player exactly 1 card so `rand.Intn(1)` always returns 0
- **Never assert on shuffled deck order**: Set up game state manually via `AddCard` instead of relying on `Reset`/`Shuffle`

## Formatting

**After editing any Go source file, always run `goimports -w` on the modified files before committing.** Use `goimports`, not `gofmt`.

## Lint

**Before committing, run `golangci-lint run ./...` and ensure no warnings or errors.**

## Run tests

```sh
go test -tags test ./...
```
