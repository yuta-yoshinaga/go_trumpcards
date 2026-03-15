# ADR-0013: Multi-stage Docker build with distroless image

## Status

Accepted

## Date

2026-02-20

## Context

再現可能なデプロイメント環境が必要だった。また、本番イメージのセキュリティと軽量化も求められた。

## Decision

マルチステージDockerビルドを採用:

1. Node.jsステージ: フロントエンドのビルド
2. Goステージ: バックエンドのビルド
3. 最終ステージ: Google distroless イメージ（ベースイメージのダイジェスト固定）

## Consequences

- 最終イメージにビルドツールが含まれず、攻撃面が最小化
- ダイジェスト固定により再現可能なビルドを保証
- distroless イメージにはシェルがないため、デバッグ時にはマルチステージの中間イメージを使用する必要がある
