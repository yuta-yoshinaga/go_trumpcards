# ADR-0008: Biome for linting and formatting (replacing ESLint)

## Status

Accepted

## Date

2026-02-21

## Context

ESLint + Prettierの2ツール構成は設定の重複や競合が発生しやすかった。より高速でシンプルな単一ツールが求められた。

## Decision

ESLintとPrettierをBiomeに統合。`npm run check` で lint とフォーマットチェックを一括実行。

## Consequences

- 単一ツールでlint + フォーマットを実行でき、設定がシンプル
- RustベースのBiomeはESLintより高速
- ESLintの一部ルールがBiomeに未対応の場合がある
- `npm run check` コマンドでCI/ローカル双方で同じチェックを実行
