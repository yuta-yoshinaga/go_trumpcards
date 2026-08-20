# Architecture Decision Records (ADR)

This directory contains Architecture Decision Records for the go_trumpcards project.

## Format

Each ADR follows the format:

- **Status**: Proposed / Accepted / Superseded / Deprecated
- **Context**: What problem or situation prompted this decision?
- **Decision**: What was decided?
- **Consequences**: What are the trade-offs and implications?

**記述言語: 日本語**（タイトル `# ADR-NNNN:` のみ英語可）。新規ADR追加時は本インデックスも更新すること。

**ADR作成基準**: 以下の3つ全てに該当する場合のみADRを作成する（詳細は [`CLAUDE.md`](../../CLAUDE.md) の「ADR記録のリトマステスト」を参照）:
1. 他の選択肢を真剣に検討した
2. 覆すと複数ファイル/レイヤーの変更が必要
3. 6ヶ月後の新メンバーが「なぜ？」と疑問に思う

ADR番号は連番ではない — 欠番はリトマステスト導入時に非アーキテクチャ的と判断され削除されたレコード。

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
| [ADR-0015](0015-accessibility-wcag.md) | WCAG accessibility compliance | Accepted | 2026-03-11 |
| [ADR-0016](0016-production-middleware.md) | Production middleware stack (CORS, security headers) | Accepted | 2026-03-01 |
| [ADR-0019](0019-ci-cd-pipeline.md) | CI/CD pipeline (CodeQL, golangci-lint, auto-tagging) | Accepted（CodeQL の PR トリガーのみ [ADR-0034](0034-codeql-post-merge.md) が置換） | 2021-04-15 |
| [ADR-0021](0021-bun-package-manager.md) | Migrate package manager from npm to bun | Accepted | 2026-03-15 |
| [ADR-0022](0022-automated-quality-gates.md) | Automated quality gates via Claude Code hooks | Accepted | 2026-03-20 |
| [ADR-0023](0023-api-documentation.md) | GoDoc/TSDoc + GitHub PagesによるAPIドキュメント自動生成 | Accepted | 2026-03-20 |
| [ADR-0026](0026-relax-coverage-target.md) | ブランチカバレッジ基準を100%から80%に緩和 | Accepted | 2026-03-23 |
| [ADR-0027](0027-cloudflare-workers-wasm.md) | Cloudflare Workers (TinyGo/Wasm) によるエッジデプロイ | Accepted | 2026-03-28 |
| [ADR-0028](0028-kv-session-persistence.md) | Cloudflare KV によるセッション永続化 | Accepted | 2026-03-28 |
| [ADR-0029](0029-design-system.md) | デザインシステム (DESIGN.md) の導入 | Accepted | 2026-04-04 |
| [ADR-0030](0030-design-tokens-only-in-source.md) | ソースコードではデザイントークンのみ使用 / 生Tailwindパレット禁止 | Accepted | 2026-04-19 |
| [ADR-0031](0031-registry-consolidation.md) | 4 レイヤーゲームレジストリの取り扱い方針 | Accepted | 2026-05-02 |
| [ADR-0032](0032-fourth-worker-capacity.md) | 4 つ目の Cloudflare Worker（容量バケット）の追加 | Accepted | 2026-06-29 |
| [ADR-0033](0033-procedural-non52-card-rendering.md) | 非52枚デッキ札の手続き的（CSS/SVG）レンダリング | Accepted | 2026-07-06 |
| [ADR-0034](0034-codeql-post-merge.md) | CodeQL を PR ではなくマージ後に実行する | Accepted | 2026-07-26 |
| [ADR-0035](0035-memory-mobile-pair-count.md) | 神経衰弱のペア数を可変にし、モバイル縦のデフォルトを下げる | Accepted | 2026-07-28 |
| [ADR-0036](0036-fifth-sixth-worker-capacity.md) | 5 つ目・6 つ目の Cloudflare Worker（容量バケット）の追加 | Accepted | 2026-07-28 |
| [ADR-0037](0037-seventh-worker-capacity.md) | 7 つ目の Cloudflare Worker（容量バケット）の追加 | Accepted | 2026-08-21 |
