# ADR-0021: Migrate package manager from npm to bun

## Status

Accepted

## Date

2026-03-15

## Context

ローカル開発時にnpm/npxのメモリ使用量が大きく、動作が重くなる問題があった。Docker imageサイズもnode:alpine (~170MB) では大きかった。全フロントエンド依存ツール（Vite, Vitest, Biome, Playwright, Tailwind, TypeScript）がbunと完全互換であることを確認済み。

## Decision

パッケージマネージャーをnpmからbun 1.3.10に移行する:

- ロックファイル: `package-lock.json` → `bun.lock`（テキスト形式）
- Docker: `node:24-alpine` → `oven/bun:1.3.10-alpine`（ダイジェスト固定）
- CI: `actions/setup-node` → `oven-sh/setup-bun@v2`（バージョン固定: 1.3.10）
- バージョン固定戦略: 完全なパッチバージョン固定（最大限の再現性）

`deploy-repomix.yml` はフロントエンドとは独立しているため、移行対象外とする。

## Consequences

- メモリ使用量の大幅削減（npm比で50-70%削減の報告あり）
- インストール速度の向上（npm比で10-25倍高速）
- Docker imageサイズの縮小（`oven/bun:alpine` ~60-70MB vs `node:alpine` ~170MB）
- bunのバージョンアップは手動で行う必要がある（パッチバージョン固定のため）
- ~~`bun test` はbun独自のテストランナーを起動するため、Vitestの実行には `bun run test` を使う必要がある~~ → ADR-0025でVitestからbun:testに移行済み
