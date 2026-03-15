# ADR-0004: React + Vite + TypeScript frontend

## Status

Accepted

## Date

2026-02-19

## Context

レガシーのHTMLページはBootstrapとjQueryで構築されており、コンポーネントの再利用が困難だった。ゲーム数の増加に伴い、モダンなフロントエンドフレームワークへの移行が必要になった。

## Decision

React + Vite + TypeScriptでフロントエンドを再構築する:

- Viteによる高速な開発サーバーとビルド
- TypeScriptによる型安全性
- 再利用可能なコンポーネント（`CardImage`、`CardBack`、`NavBar`等）
- SPAとしてGo Webサーバーから静的ファイルを配信

## Consequences

- コンポーネントベースで再利用性が大幅に向上
- TypeScriptによるコンパイル時の型チェックでバグを早期発見
- Viteにより開発時のHMRが高速
- Node.js/npmが新たな依存として必要
- ビルド成果物を `public/` に配置し、GoサーバーからSPAとして配信
