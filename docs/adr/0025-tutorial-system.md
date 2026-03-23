# ADR-0025: インタラクティブチュートリアルシステム

## Status

Accepted

## Date

2026-03-23

## Context

トランプゲームのルールや操作方法が分からない初心者ユーザーが、実際の画面を見ながら学べる仕組みがなかった。既存のゲームマニュアル（`docs/manual/`）は静的ドキュメントであり、実際の画面操作と結びついていない。19ゲーム全てに統一的な仕組みで提供する必要がある。

外部ライブラリ（react-joyride、driver.js等）の採用も検討したが、以下の理由からカスタム実装を選択した:

- プロジェクトはUIライブラリ不使用で、全コンポーネントがTailwind + カスタム実装
- ゲームの状態（phase）やAPIコールとの連動が必要で、汎用ライブラリでは対応しにくい
- 既存の`ConfirmDialog`、`ActionLogPanel`のオーバーレイパターン（z-50、フォーカストラップ、glass-panel）を再利用できる
- framer-motionによるアニメーション基盤が既にある
- バンドルサイズの増加を最小限にできる

## Decision

外部ライブラリを使用せず、以下のカスタムコンポーネント群でチュートリアルシステムを構築する:

### アーキテクチャ

1. **`TutorialProvider`** — Contextベースのプロバイダー。各ゲームページ内に配置し（App.tsxではなく）、ゲームごとに独立したチュートリアル設定を持つ。アクティブ時にオーバーレイを自動レンダリング。

2. **`TutorialOverlay`** — フルスクリーンオーバーレイ。SVG `<mask>` でスポットライトのくり抜きを実現し、`ResizeObserver`でリサイズに追従。`ConfirmDialog`から`getFocusableElements`を再利用したフォーカストラップを実装。

3. **`TutorialTooltip`** — glass-panelスタイルのツールチップ。ステップインジケーター（`1/5`）、「次へ」「スキップ」ボタンを表示。

4. **`useTutorial`フック** — ステップ遷移、完了状態の`localStorage`永続化、`onEnter`コールバックを管理。

5. **`data-tutorial`属性** — ゲームページの各操作要素にターゲット指定用の属性を付与。CSSセレクタ依存を避け、リファクタリングに強い設計。

### ステップ進行モデル

- `advanceOn: 'next'` — ツールチップの「次へ」ボタンで手動進行
- `advanceOn: 'click'` — 対象要素のクリックイベントを監視し、クリック後に自動で次ステップへ
- キーボード操作: Enter で次へ、Escape でスキップ

### i18n対応

- 共通テキスト（次へ、スキップ、完了）は`tutorial`名前空間
- ゲーム固有のステップ説明は各ゲームの既存名前空間内に`tutorial`キーとして追加
- `TutorialProvider`の`translateMessage`プロップでゲーム名前空間の`t()`を渡す

### 段階的展開

- Phase 1: 基盤コンポーネント + BlackJack（本ADR）
- Phase 2: 全ゲームへの展開（Issue #860）
- Phase 3: ヒント表示・自動提案・進捗管理（Issue #861）

## Consequences

### メリット

- 外部依存なし — バンドルサイズ増加は最小限（約2KB gzip）
- 既存UIパターン（glass-panel、フォーカストラップ、framer-motion）と統一的
- ゲーム状態との連動が自由に制御可能
- `data-tutorial`属性によるターゲット指定で、HTML構造の変更に強い
- `localStorage`による完了状態永続化で、再訪時の不要な表示を防止

### デメリット

- 汎用ライブラリと比較して、高度な機能（スクロール追従、複数ハイライト等）は自前実装が必要
- 新ゲーム追加時にチュートリアルステップ定義の手動作成が必要（自動生成なし）
- SVG maskアプローチはブラウザ間の描画差異に注意が必要（主要ブラウザでは問題なし）

### アクセシビリティ

- `role="dialog"` + `aria-modal="true"` でスクリーンリーダー対応
- `aria-live="polite"` でステップ変更の通知
- フォーカストラップ（`ConfirmDialog`と同一パターン）
- `prefers-reduced-motion`対応（`useReducedMotion`フック利用）
