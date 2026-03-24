# ADR-0027: Video Poker (Jacks or Better) の追加

## Status

Accepted

## Date

2026-03-24

## Context

Issue #867 にて、1人用カジノゲームの追加が要望された。既存のポーカー系ゲーム（5-card Draw Poker、Texas Hold'em、Omaha Hold'em、Indian Poker）はすべてマルチプレイヤー対戦型であり、1人でカジュアルに遊べるカジノスタイルのポーカーゲームが不足していた。

Video Poker (Jacks or Better) は以下の理由で候補として最適であった:

1. **既存インフラの再利用**: `hand_eval.go` の `evalFiveCardHand` 関数がそのまま手札評価に使える
2. **シンプルなゲームフロー**: Bet → Draw → Result の3フェーズモデルで、Baccarat と同様の単純なライフサイクル
3. **ChipHolder の再利用**: 既存のチップ/ベッティングシステムをそのまま利用可能
4. **CPU AI 不要**: 1人用ゲームのため、CPU 戦略ロジックの実装が不要

## Decision

1. **evalFiveCardHand の直接再利用**: `hand_eval.go` と同一パッケージ（`domain`）に VideoPoker を実装し、既存の手札評価関数をインポートなしで直接呼び出す
2. **Jacks or Better 判定関数の新規実装**: `evalFiveCardHand` の結果に対して、J 以上のワンペアを最低配当条件とする判定ロジックを別関数として実装
3. **3フェーズモデル**: Bet (0) → Draw (1) → Result (2) のシンプルなフェーズ遷移を採用。Baccarat (Bet → End) よりも1フェーズ多いが、ホールド選択のためのインタラクションフェーズが必要
4. **1-5 コインベットシステム**: 伝統的な Video Poker のベットシステムを採用。5コイン時のロイヤルフラッシュ配当を 4000 に設定（標準的な Jacks or Better ペイテーブル）

## Consequences

- **最小限の新規コード**: 手札評価・チップ管理・アクションログなど既存インフラを最大限活用し、新規実装はゲーム固有ロジック（ベット処理、ホールド/ドロー、配当計算、Jacks or Better 判定）に限定される
- **パターンの一貫性**: Baccarat と同様の1人用カジノゲームパターンに従い、既存のアーキテクチャとの整合性を維持
- **将来の拡張性**: Deuces Wild や Joker Poker など他の Video Poker バリアントへの拡張が、ペイテーブルと判定関数の差し替えのみで実現可能
