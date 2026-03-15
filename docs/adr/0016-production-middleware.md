# ADR-0016: Production middleware stack (CORS, security headers)

## Status

Accepted

## Date

2026-03-01

## Context

Web APIを本番環境で公開するにあたり、セキュリティヘッダーとCORS設定が必要だった。

## Decision

- **CORS**: `CORS_ALLOWED_ORIGINS` 環境変数で設定可能。本番で未設定の場合はスキップ
- **`DefaultProdStack`**: 情報漏洩を防ぐセキュリティヘッダーを付与するミドルウェアスタック

## Consequences

- 環境変数ベースの設定により、開発・本番で柔軟にCORS制御可能
- デフォルトで安全な設定（未設定時はCORSなし）
- セキュリティヘッダーが自動的に全レスポンスに付与
