# ADR-0014: Shared UI component and hook extraction

## Status

Accepted

## Date

2026-03-02

## Context

ゲーム数が11に増加するにつれ、各ゲームページで大量の重複コードが発生していた。カード表示、メッセージ表示、ゲーム操作のリプレイ、カード選択等のロジックが各ページに散在。

## Decision

共通コンポーネントと共通フックを体系的に抽出:

**共通コンポーネント:**
- `CardImage` / `CardBack` — カード表示
- `GameMessageBox` / `GameFooter` — メッセージ・フッター
- `ActionLogSection` — アクションログ
- `PhaseIndicator` — フェーズ表示
- `ConfirmDialog` — 確認ダイアログ
- `LoadingSpinner` — ローディング
- `SettingsPanel` — 設定パネル

**共通フック:**
- `useGameApi` — API通信
- `useCardSelection` — カード選択
- `useGameReplay` — CPUアクションリプレイ

## Consequences

- 各ゲームページのボイラープレートが大幅削減
- UIの一貫性が自動的に保証される
- 共通コンポーネントの変更が全ゲームに影響するため、テストが重要
- 新しいゲーム追加時に共通コンポーネントを組み合わせるだけで基本UIが構築可能
