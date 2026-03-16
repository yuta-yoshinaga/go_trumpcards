# ADR-0020: Domain helper extraction and DRY refactoring

## Status

Accepted

## Date

2026-03-14

## Context

11ゲームに成長する中で、複数のゲームで同一ロジックの重複が発生していた。ジョーカー判定、メモリ減衰、整数クランプ、CUIエラーメッセージ等が各ゲームに散在。

## Decision

共通ロジックを共有ヘルパーに抽出:

- **`IsJoker`**: ジョーカー判定を統一ヘルパーに集約
- **`DecayMemories`**: Memory/Doubt 共通のメモリ減衰ジェネリック関数
- **`memoryManager[T]`**: `ResetMemory`/`DecayMemories`/`AddMemory` を持つ埋め込みジェネリック構造体。`DoubtPlayer` と `MemoryPlayer` に埋め込んで状態と振る舞いを共有
- **`ClampIntPtr`**: コントローラーの `ToConfig` で使用する整数クランプヘルパー
- **`cuiutil`**: CUI共通ユーティリティパッケージ
- **`cuimsg`**: CUIエラーメッセージの一元管理

## Consequences

- コード重複が排除され、修正時の変更箇所が一箇所に集約
- ジェネリック関数により型安全な共通処理が実現
- 新しいゲーム追加時にヘルパーを再利用可能
- 過度な抽象化のリスクがあるため、3箇所以上で使われる場合のみ抽出
