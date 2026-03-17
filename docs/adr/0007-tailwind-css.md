# ADR-0007: Tailwind CSS (replacing Bootstrap)

## Status

Accepted

## Date

2026-02-21

## Context

BootstrapはReactコンポーネントパターンとの相性が悪く、不要なCSSが多かった。ユーティリティファーストのCSSフレームワークに移行することで、コンポーネント単位でスタイリングを完結させたかった。

## Decision

BootstrapをTailwind CSSに置換する。レガシーHTMLページとBootstrapの静的アセットをすべて削除。

## Consequences

- ユーティリティクラスによりコンポーネント単位でスタイルが完結
- 未使用CSSがビルド時にパージされ、バンドルサイズ削減
- Bootstrapのグリッドシステムやコンポーネントが使えなくなる
- クラス名が長くなる傾向があるが、Reactコンポーネントの再利用で緩和
