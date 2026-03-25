# ADR-0028: Video Poker バリアント（Deuces Wild・Joker Poker）の追加

## Status

Accepted

## Date

2026-03-25

## Context

Issue #903 にて、既存の Video Poker (Jacks or Better) に加えてワイルドカードを使ったバリアントの追加が要望された。具体的には以下の2つ:

1. **Deuces Wild**: 52枚デッキ、すべての2がワイルドカード
2. **Joker Poker**: 53枚デッキ（標準52枚+ジョーカー1枚）、ジョーカーがワイルドカード

各バリアントはワイルドカードの種類、デッキ構成、配当表（ペイテーブル）、最低配当条件が異なるが、ゲームフロー（Bet → Draw → Result）は共通である。

検討した設計方針:

1. **各バリアントを独立したドメインエンティティとして実装**: コードの重複が大きく、保守コスト増
2. **VideoPoker に条件分岐を追加**: 既存コードが肥大化し、新バリアント追加のたびに修正箇所が増加
3. **Strategy パターンによる差分注入**: バリアント固有のロジック（ワイルド判定、手札評価、ペイテーブル、最低配当条件）を `VideoPokerVariantConfig` として抽出し、VideoPoker ドメインに注入

## Decision

**Strategy パターン（`VideoPokerVariantConfig`）** を採用する。

1. **`VideoPokerVariantConfig` 構造体の導入**: バリアント固有のロジック（デッキサイズ、ジョーカー使用有無、ワイルド判定関数、手札評価関数、ペイテーブル関数、最低配当条件関数）をまとめた構成体を定義
2. **既存 VideoPoker ドメインの拡張**: `VideoPoker` が `VideoPokerVariantConfig` をオプショナルに受け取り、設定がある場合はバリアント固有のロジックを使用、ない場合は従来の Jacks or Better ロジックをそのまま維持（後方互換性）
3. **ワイルドカード対応の手札評価**: `evalWildHand` 関数を新規実装し、ワイルドカードの全組み合わせ置換による最適手札評価を実現
4. **エンドポイントの追加**: `/deuceswild/exec` と `/jokerpoker/exec` を新規追加。リクエスト/レスポンススキーマは VideoPoker と共通（レスポンスに `variantName` フィールドを追加）
5. **フロントエンドの共通化**: `VideoPokerGameContent` コンポーネントを共通コンテンツとして抽出し、`DeucesWildPage` と `JokerPokerPage` がこれを再利用

## Consequences

### メリット

- **最小限の新規コード**: ゲームフロー、チップ管理、アクションログなどの共通ロジックはすべて既存 VideoPoker をそのまま再利用
- **高い拡張性**: 新しい Video Poker バリアント（Bonus Poker、Double Bonus など）を追加する場合、`VideoPokerVariantConfig` を定義するだけで対応可能
- **後方互換性**: 既存の Jacks or Better のコードパスに変更なし
- **テストの効率化**: バリアント固有のロジック（ワイルド評価、ペイテーブル）のみを重点的にテストすれば良い

### デメリット

- **関数型の設定**: Go の構造体に関数フィールドを持たせるため、設定の直列化やデバッグがやや複雑
- **エンドポイント数の増加**: 24 → 26 エンドポイントに増加（ルーティングとコントローラーの登録が必要）
