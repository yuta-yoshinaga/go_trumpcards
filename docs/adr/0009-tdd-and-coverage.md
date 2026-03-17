# ADR-0009: TDD cycle and 100% branch coverage requirement

## Status

Accepted

## Date

2026-02-19

## Context

カードゲームはランダム性が高く、バグの検出が困難。確実な品質保証のため、厳格なテスト方針が必要だった。

## Decision

1. **TDDサイクル（Red-Green-Refactor）の必須化**: すべての実装はテストを先に書く
2. **100%ブランチカバレッジ（C1）**: `cmd/` と `internal/infrastructure/` を除くすべての `internal/` パッケージ、フロントエンドの `api/`・`components/`・`pages/`・`utils/` で必須
3. **4層テスト構造**: Domain、Use Case、Presenter、Controller の各レイヤーでテスト
4. **決定的テスト**: `AddCard` による手動セットアップ、シャッフル順序に依存しない、ランダム分岐はリトライループ（最大1000回）で両分岐をカバー
5. **モックパターン**: `testify/mock` を使用、`*_mock.go` ファイルをインターフェース隣接に配置

## Consequences

- 高い品質保証: 11ゲームすべてで回帰テストが機能
- ランダム性に依存しないテストにより、CIでのフレイキーテストを排除
- テスト作成コストが高いが、バグ修正コストが大幅に削減
- 新しいゲーム追加時のテストテンプレートが確立されている
