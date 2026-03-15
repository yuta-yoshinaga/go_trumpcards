# ADR-0002: Presenter pattern for output abstraction

## Status

Accepted

## Date

2020-11-29

## Context

Clean Architecture においてインタラクタ（ユースケース層）が出力フォーマットを直接扱うと、CUIとWebで別々のインタラクタが必要になる。出力処理をビジネスロジックから分離する仕組みが必要だった。

## Decision

Presenterパターンを採用する:

- `internal/usecase/presenter/` にプレゼンターインターフェースを定義
- `internal/adapter/presenter/` にCUI用・Web用の具体実装を配置
- インタラクタにプレゼンターをDIし、出力形式をインタラクタから隠蔽

例: `BlackJackPresenter` インターフェースを `BlackJackCuiPresenter`（テキスト出力）と `BlackJackWebPresenter`（JSON出力）が実装。

## Consequences

- インタラクタのコードを変更せずに新しい出力形式（Web JSON、CUIテキスト等）を追加可能
- テスト時はモックプレゼンター（`*_mock.go`）を注入してI/Oなしでテスト可能
- 各ゲームごとにプレゼンターインターフェース・CUI実装・Web実装の3ファイルが必要
- 11ゲーム × 3ファイル = 33ファイルとなるが、パターンが統一されているため保守負荷は低い
