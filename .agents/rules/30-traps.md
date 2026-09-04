---
trigger: always_on
---
# Pitfalls & Traps

このリポジトリで実際に事故になった罠と対策。

- **i18n の `ns:key` は解決しない**:
  - 症状: 画面に `foo:bar` 等の生識別子が出る（エラーも警告も出ない）。
  - 原因: i18next の既定 `nsSeparator` が `:` のため。
  - どうするか: キーの形は `hint.<snake_case>` のように同一 ns 内のドット区切りにする。
- **共有ページでの i18n 名前空間**:
  - 症状: 片方の JSON にキーを足しただけでは別モードで文言が出ない。
  - 原因: 複数モードで 1 ページ共有時、`t` は `useGamePageSetup(gameKey)` 経由でモードごとに別 ns を引くため。
  - どうするか: 共有される全モードの JSON に翻訳キーを追加する。
- **`api/openapi.yaml` は唯一 CRLF**:
  - 症状: 編集後に差分が数万行になる、または折り畳みスカラー (`>` `|`) 破損でパースに失敗する。
  - 原因: このリポジトリで唯一 CRLF 改行を採用しているため。
  - どうするか: CRLF を維持して編集し、編集後は必ず差分行数を確認する。
- **レスポンスと OpenAPI の同期**:
  - 症状: CI のスキーマ突合で失敗する。
  - 原因: Go 側のレスポンスにフィールドを追加したが OpenAPI 側に未反映。
  - どうするか: レスポンスに新フィールドを足したら `api/openapi.yaml` にも必ず足す。
- **Cloudflare Worker のビルドタグ**:
  - 症状: `go build ./...` が通るのに Worker ビルドや CI で失敗する。
  - 原因: 全ファイルが `!js || !wasm` で stranded symbol を通常ビルドで検出できないため。
  - どうするか: `GOOS=js GOARCH=wasm go build -tags <worker> -o /dev/null ./cmd/workers/<worker>` で型検査する。
- **TinyGo での `errors.As`**:
  - 症状: Worker 実行時やテスト時に panic が発生する。
  - 原因: TinyGo では reflect の動的 `AssignableTo` が未実装。
  - どうするか: Worker バイナリに入るコードで `errors.As` を使わない（`_test.go` は Worker に入らないので対象外）。
- **非公開フィールドだけの構造体**:
  - 症状: JSON シリアライズ結果が `{}` になる。
  - 原因: 非公開（小文字）フィールドは標準エンコーダで無視されるため。
  - どうするか: フィールドを公開するか、`MarshalJSON` を実装する。
- **ゲーム追加・改名時の登録点**:
  - 症状: CI の件数アサーションやドキュメント整合性チェックが失敗する。
  - 原因: 登録点が 5 箇所あり、`registry_test.go` の件数や README / `docs/games.md` / `api/openapi.yaml` の `tags:` と CI で突き合わされるため。
  - どうするか: [`docs/new-game-checklist.md`](../../docs/new-game-checklist.md) の手順を漏れなく完了させる。
