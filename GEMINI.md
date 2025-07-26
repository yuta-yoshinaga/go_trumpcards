# GEMINI.md

This file provides guidance to GEMINI when working with code in this repository.

## Repository Overview

This repository contains a Go implementation of a trump card game algorithm. The project is structured following the principles of Clean Architecture. The main application can be run as a command-line user interface (CUI).

## Architecture

The codebase follows **Clean Architecture**, as mentioned in the `README.md`. This means the code is organized into distinct layers with a strict dependency rule: outer layers can depend on inner layers, but inner layers cannot know anything about outer layers. The typical layers are:

- **Entities**: Core business objects (e.g., Card, Deck, Player).
- **Use Cases (Interactors)**: Application-specific business rules.
- **Interface Adapters**: Controllers, gateways, and presenters that convert data for use cases and UI.
- **Frameworks & Drivers**: The outermost layer, including the UI (like the CUI), database, and web frameworks.

When adding new features, ensure that code is placed in the appropriate layer and that the dependency rule is respected.

## Common Commands

- **Run the application (CUI):**
  ```sh
  go run main.go cui
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