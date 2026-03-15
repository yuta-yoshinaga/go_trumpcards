# ADR-0001: Clean Architecture adoption

## Status

Accepted

## Date

2020-11-29

## Context

トランプゲームのアルゴリズムを実装するGoプロジェクトを開始するにあたり、ビジネスロジックをUI・インフラから分離し、テスタビリティと拡張性を確保する必要があった。ゲームが増えてもコアロジックに影響を与えずにCLI・Web等のUIを追加できる構造が求められた。

## Decision

Clean Architectureを採用し、以下の4層構造とする:

1. **Domain** (`internal/domain/`): コアビジネスロジック（カード、デッキ、ゲームルール）
2. **Use Case** (`internal/usecase/`): アプリケーションビジネスルール（インタラクタ + プレゼンターインターフェース）
3. **Adapter** (`internal/adapter/`): データ変換層（コントローラ + プレゼンター実装）
4. **Infrastructure** (`internal/infrastructure/`): 外部接続（CLI、Webサーバー）

依存方向は外側から内側への一方向のみ（`infrastructure` -> `adapter` -> `usecase` -> `domain`）。

## Consequences

- ゲームロジックがUIから完全に独立し、CUI・Web双方に同じロジックを再利用できる
- 各レイヤーが独立してテスト可能（モックによるユニットテスト）
- 新しいゲーム追加時は各レイヤーにファイルを追加するだけで済む
- レイヤー間のインターフェース定義が必要なため、初期の実装コストは高い
- 11ゲームまでスケールした現在も構造が維持されている
