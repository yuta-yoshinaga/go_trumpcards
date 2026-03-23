# Architecture Decision Records (ADR)

This directory contains Architecture Decision Records for the go_trumpcards project.

## Format

Each ADR follows the format:

- **Status**: Accepted / Superseded / Deprecated
- **Context**: What problem or situation prompted this decision?
- **Decision**: What was decided?
- **Consequences**: What are the trade-offs and implications?

**記述言語: 日本語**（タイトル `# ADR-NNNN:` のみ英語可）。新規ADR追加時は本インデックスも更新すること。

## Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-0001](0001-clean-architecture.md) | Clean Architecture adoption | Accepted | 2020-11-29 |
| [ADR-0002](0002-presenter-pattern.md) | Presenter pattern for output abstraction | Accepted | 2020-11-29 |
| [ADR-0003](0003-golang-standards-layout.md) | golang-standards/project-layout directory structure | Accepted | 2026-02-23 |
| [ADR-0004](0004-react-frontend.md) | React + Vite + TypeScript frontend | Accepted | 2026-02-19 |
| [ADR-0005](0005-stateless-rest-api.md) | Stateless REST API with session-based state isolation | Accepted | 2026-02-20 |
| [ADR-0006](0006-openapi-specification.md) | OpenAPI specification as API contract | Accepted | 2026-02-21 |
| [ADR-0007](0007-tailwind-css.md) | Tailwind CSS (replacing Bootstrap) | Accepted | 2026-02-21 |
| [ADR-0008](0008-biome-linter.md) | Biome for linting and formatting (replacing ESLint) | Accepted | 2026-02-21 |
| [ADR-0009](0009-tdd-and-coverage.md) | TDD cycle and 100% branch coverage requirement | Superseded by ADR-0026 | 2026-02-19 |
| [ADR-0010](0010-tanstack-react-query.md) | TanStack React Query for API state management | Accepted | 2026-03-04 |
| [ADR-0011](0011-i18n-react-i18next.md) | i18n with react-i18next and browser language detection | Accepted | 2026-03-05 |
| [ADR-0012](0012-playwright-e2e.md) | Playwright for E2E testing | Accepted | 2026-03-04 |
| [ADR-0013](0013-docker-distroless.md) | Multi-stage Docker build with distroless image | Accepted | 2026-02-20 |
| [ADR-0014](0014-shared-ui-components.md) | Shared UI component and hook extraction | Accepted | 2026-03-02 |
| [ADR-0015](0015-accessibility-wcag.md) | WCAG accessibility compliance | Accepted | 2026-03-11 |
| [ADR-0016](0016-production-middleware.md) | Production middleware stack (CORS, security headers) | Accepted | 2026-03-01 |
| [ADR-0017](0017-interactive-cli-mode.md) | Interactive CLI mode with game switching | Accepted | 2026-03-11 |
| [ADR-0018](0018-ansi-color-output.md) | ANSI color output for CLI with TTY auto-detection | Accepted | 2026-03-14 |
| [ADR-0019](0019-ci-cd-pipeline.md) | CI/CD pipeline (CodeQL, golangci-lint, auto-tagging) | Accepted | 2021-04-15 |
| [ADR-0020](0020-domain-helper-extraction.md) | Domain helper extraction and DRY refactoring | Accepted | 2026-03-14 |
| [ADR-0021](0021-bun-package-manager.md) | Migrate package manager from npm to bun | Accepted | 2026-03-15 |
| [ADR-0022](0022-automated-quality-gates.md) | Automated quality gates via Claude Code hooks | Accepted | 2026-03-20 |
| [ADR-0023](0023-api-documentation.md) | GoDoc/TSDoc + GitHub PagesによるAPIドキュメント自動生成 | Accepted | 2026-03-20 |
| [ADR-0024](0024-fluid-tactile-ui-redesign.md) | Fluid & Tactile UIリデザイン | Accepted | 2026-03-21 |
| [ADR-0025](0025-tutorial-system.md) | インタラクティブチュートリアルシステム | Accepted | 2026-03-23 |
| [ADR-0026](0026-relax-coverage-target.md) | ブランチカバレッジ基準を100%から80%に緩和 | Accepted | 2026-03-23 |
| [ADR-0027](0027-video-poker.md) | Video Poker (Jacks or Better) の追加 | Accepted | 2026-03-24 |
