# ADR-0003: golang-standards/project-layout directory structure

## Status

Accepted

## Date

2026-02-23

## Context

ゲーム数が増加するにつれ、フラットなディレクトリ構造ではファイルの管理が困難になった。Goコミュニティの標準的なプロジェクトレイアウトに合わせることで、新規参加者にとっても理解しやすい構造にする必要があった。

## Decision

`golang-standards/project-layout` に従い、以下のようにリストラクチャリング:

- `cmd/` — エントリポイント（CLI、サーバー）
- `internal/` — ビジネスロジック全体（Clean Architecture レイヤーで構成）
- `api/` — OpenAPI仕様
- `frontend/` — Reactフロントエンド

これは破壊的変更（`refactor!`）としてコミットされた。

## Consequences

- Goの標準的なプロジェクト構造に準拠し、可読性向上
- `internal/` によりパッケージの外部公開を防止
- 既存のインポートパスすべてが変更となった（破壊的変更）
- 以降のすべてのゲーム追加がこの構造に従っている
