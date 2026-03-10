---
globs: ["**/*.go"]
---

# Go ファイル編集ルール

## フォーマット

- **`goimports -w <file>` を必ず実行してからコミットする**（`gofmt` は使わない）
- パス: `/home/yuta/go/bin/goimports`

## Lint

- **コミット前に `golangci-lint run ./...` を実行し、警告・エラーがないことを確認する**

## テスト

- **ユニットテストは必須**。実装と同じコミットに含める
- **TDDサイクル (Red → Green → Refactor)** に従う
- テスト実行: `go test -tags test ./...`

### カバレッジ基準

- `cmd/` と `internal/infrastructure/` はカバレッジ対象外
- それ以外の `internal/` 配下は **ブランチカバレッジ (C1) 100%** が必須
- if/else・switch・ループ終了条件など、すべての分岐を網羅すること

### テスト配置（レイヤー別）

| レイヤー | テストファイル |
|---------|--------------|
| Domain | `internal/domain/*_test.go` |
| Use cases | `internal/usecase/*Interactor_test.go` |
| Presenters | `internal/adapter/presenter/*_test.go` |
| Controllers | `internal/adapter/controller/*_test.go` |

### モックパターン

- **Presenterモック**: `internal/usecase/presenter/*_mock.go` — `testify/mock` で実装
- **Interactorモック**: `internal/adapter/controller/usecase/*_mock.go` — `testify/mock` で実装
- 既存の `BlackJack*_mock.go` を参照パターンとして使う

### 決定的テストの書き方

シャッフルに依存しないようにすること:

- `AddCard` でハンドを手動セットアップし、`Reset`/`Shuffle` 後の順序に依存しない
- ディーラー/CPUに自動ドローが発生しないスコアを与える（例: BlackJack dealer >= 17）
- ランダム分岐のカバレッジは最大1000回のリトライループで両分岐を確保する

## アーキテクチャ

Clean Architecture: `infrastructure` → `adapter` → `usecase` → `domain`

- ドメインインターフェース: `internal/domain/interfaces/`
- 依存方向を逆にしてはいけない（外側のレイヤーは内側に依存する）

## デッドコード

- コード変更時に遭遇したデッドコードは必ず削除する
- 検出ツール: `golang.org/x/tools/cmd/deadcode`
- 削除前に手動で確認する（リフレクション経由の呼び出しなどの誤検知に注意）
