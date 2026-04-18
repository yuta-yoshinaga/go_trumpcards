# ADR-0030: ソースコードではデザイントークンのみ使用 / 生Tailwindパレット禁止

## Status

Accepted

## Date

2026-04-19

## Context

ADR-0029で`DESIGN.md`とデザイントークン (`--color-ds-*`) を導入したが、採用当時は既存コードの段階的移行方針であったため、`text-white/70`・`bg-green-700/50`・`ring-yellow-400`・`text-red-400` など生のTailwindパレット指定がページ/コンポーネント全体に約240箇所残存していた（issue #1411）。

この残存は以下の実害を生じていた:

- **アクセシビリティ**: `text-white/50` は背景 `--color-ds-bg` に対して約3.9:1で、WCAG AA（通常テキスト4.5:1）に未達。一方 `--color-ds-text-muted` は約6.2:1を保証する。
- **ブランド**: DESIGN.mdの "Restrained — one gold accent" 方針に反し、赤/緑/青/黄/アンバーが個別ページで混在。
- **ライトモード移行の阻害**: DESIGN.mdが規定するライトモード（`--bg: #F5F0E8` など）に切り替えた際、`text-white/N` や `bg-green-700` のようなモード非依存のリテラルはそのままでは破綻する。

以下の代替案を検討した:

1. **現状維持（段階的移行を継続）** — 目視レビューのみに頼るため再発が止まらない。過去数ヶ月で件数が増加傾向。
2. **Biomeカスタムルール (`no-restricted-syntax` 相当)** — Biome 2.4.8には相当する汎用ルールが未搭載。experimental pluginsは将来的な手段として有力だが現時点では自前のJSXパース実装が必要。
3. **ユニットテスト化（vitestで各ファイルをread）** — 過剰。テスト実行時間を増やす割にCIの他処理と分離しづらい。
4. **独立したNode/Bunガードスクリプトを `bun run check` に組み込む（採用）** — 依存なし・即時エラー・数百ms。Biomeが対応したら差し替え可能。

## Decision

1. `frontend/src/**/*.{ts,tsx}`（`*.test.*` を除く）では以下をビルド失敗とする:
   - `text-white/\d+`（opacity付きの白テキスト）
   - 生Tailwindパレット: `(bg|text|border|ring|shadow|from|to|via|divide)-(red|green|yellow|amber|blue|orange|purple|pink|emerald|sky|indigo|violet|fuchsia|rose|slate|zinc|gray|neutral|stone|cyan|teal|lime)-\d{2,3}(/\d+)?`
2. 代わりに `frontend/src/index.css` の `@theme` に定義されている `ds-*` トークン（`text-ds-text-primary`、`bg-ds-error`、`ring-ds-warning` 等）を使う。
3. ガードスクリプトは `frontend/scripts/check-design-tokens.mjs` として実装し、`bun run check` の最後に走らせる（Biomeのチェック後）。
4. テストファイル（`*.test.ts(x)`）は対象外。テストがUIクラスをアサーションする際は、アサーション文字列もトークン名に追従させる。

### セマンティックマッピング

| 生パレット | デザイントークン |
|----------|----------------|
| `red-*` | `ds-error` |
| `green-*` | `ds-success` |
| `yellow/amber-*` | `ds-warning` |
| `blue-*` | `ds-info` |
| `orange-*` | `ds-warning/80` |
| `gray-{500..900}` | `ds-surface-elevated`（bg）/ `ds-text-muted`（text）/ `ds-border-subtle`（border） |
| `text-white/80..90` | `text-ds-text-primary` |
| `text-white/50..70` | `text-ds-text-muted` |
| 金色ボタン上の暗色テキスト | `text-ds-text-on-accent` |
| 明色サーフェス（白フィールド等）上の暗色テキスト | `text-ds-text-inverse` |

## Consequences

### ポジティブ
- 新規/修正コードでデザインシステム外しが原理的に発生しない（CIで落ちる）。
- `--color-ds-*` 一箇所の更新で全UIの色を変更でき、ライトモード実装時の書き換えが最小になる。
- WCAG AA未達のまま放置される"白テキスト+半透明"が排除される。
- コードレビューで「この色なぜここだけ別？」という議論が消える。

### ネガティブ
- Biome設定ではなく独自スクリプトが増える（将来Biomeが対応したら差し替え）。
- ゲーム固有の色（例: daifugoの`--color-daifugo-revolution`）は本ルールの対象外 — `@theme` に `--color-*` として定義済みなら利用可。新規の一回限りの色を追加したい場合は、まず`@theme`に追加してから使う運用とする。
- 既存のテストアサーションでクラス名を参照しているものは同時更新が必要になる（本PRで対応済み）。
