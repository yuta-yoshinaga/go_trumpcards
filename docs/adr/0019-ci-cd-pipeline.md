# ADR-0019: CI/CD pipeline (CodeQL, golangci-lint, auto-tagging)

## Status

Accepted

## Date

2021-04-15

## Context

コード品質とセキュリティの自動チェック、およびリリースの自動化が必要だった。

## Decision

GitHub Actionsで以下のパイプラインを構築:

- **CodeQL**: push/PR時にセキュリティスキャン（PR時の実行は [ADR-0034](0034-codeql-post-merge.md) で廃止。現在は `develop` への push と週次のみ）
- **golangci-lint**: Go コードの静的解析
- **CI**: バックエンド・フロントエンド双方のテスト自動実行（2026-02-19追加）
- **Auto-tag**: `master` へのマージ時に自動でgitタグとGitHub Releaseを作成

## Consequences

- セキュリティ脆弱性とコード品質問題を自動検出
- リリースプロセスが自動化され、手動タグ付けが不要
- PRマージ前に全テストが通ることを保証
