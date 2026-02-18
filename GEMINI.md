# GEMINI.md

This file provides guidance to GEMINI when working with code in this repository.

## Repository Overview

This repository contains a Go implementation of trump card game algorithms. The project is structured following the principles of Clean Architecture. The following games are implemented:

- **BlackJack**: CLI and Web GUI
- **Poker (5-card Draw)**: CLI and Web GUI
- **Old Maid (Babanuki)**: CLI and Web GUI

## Architecture

The codebase follows **Clean Architecture**, as mentioned in the `README.md`. This means the code is organized into distinct layers with a strict dependency rule: outer layers can depend on inner layers, but inner layers cannot know anything about outer layers. The typical layers are:

- **Entities**: Core business objects (e.g., Card, Deck, Player).
- **Use Cases (Interactors)**: Application-specific business rules.
- **Interface Adapters**: Controllers, gateways, and presenters that convert data for use cases and UI.
- **Frameworks & Drivers**: The outermost layer, including the UI (like the CUI), database, and web frameworks.

When adding new features, ensure that code is placed in the appropriate layer and that the dependency rule is respected.

## Common Commands

- **Run the application:**
  ```sh
  go run main.go cui      # BlackJack CLI
  go run main.go poker    # 5-card Draw Poker CLI
  go run main.go oldmaid  # Old Maid CLI
  go run main.go web      # Start REST API + web GUI server
  ```

- **Run all tests:**
  The project uses the `testify` package for assertions.
  ```sh
  go test ./...
  ```

- **Manage dependencies:**
  To ensure the `go.mod` and `go.sum` files are up-to-date.
  ```sh
  go mod tidy
  ```

## Git Workflow & CI/CD

- **Development Branch:** `develop`. All pull requests should target this branch. CodeQL analysis is run on pushes and PRs to `develop`.
- **Release Branch:** `master`. A push to `master` automatically triggers a workflow that bumps the version, creates a git tag, and generates a GitHub Release.

## Commit Message Format

All commit messages must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**

| Type | Description |
|------|-------------|
| `feat` | A new feature |
| `fix` | A bug fix |
| `docs` | Documentation only changes |
| `style` | Changes that do not affect code meaning (formatting, etc.) |
| `refactor` | A code change that neither fixes a bug nor adds a feature |
| `perf` | A code change that improves performance |
| `test` | Adding missing tests or correcting existing tests |
| `chore` | Changes to the build process or auxiliary tools |

**Examples:**

```
feat(entities): add new card type to BlackJack
fix(poker): correct hand ranking for flush detection
docs: update README with web deployment instructions
test(blackjack): add tests for dealer bust scenario
refactor(usecases): simplify interactor dependency injection
```

**Rules:**
- The description must be in lowercase and not end with a period.
- Use the imperative mood in the description (e.g., "add feature" not "added feature").
- Breaking changes must include `BREAKING CHANGE:` in the footer or append `!` after the type/scope.