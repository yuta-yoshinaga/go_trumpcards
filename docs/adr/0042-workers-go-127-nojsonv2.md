# ADR-0042: Worker の Go 1.27 更新と encoding/json v1 の維持

## Status

Accepted

## Date

2026-09-24

## Context

Go 1.27 が公開され、Go 1.25 はセキュリティ修正の対象外になった。Go 1.27 の `encoding/json` v2 は Worker を gzip 1,048,576 B の上限より約 140 KB大きくするため、Worker ビルドには `GOEXPERIMENT=nojsonv2` が必要となる。`go.mod` の `go 1.25.8` は変更しない。

PR #8043 で staging の全 Worker が `runtime.getRandomData` の不足による WebAssembly LinkError で起動しなかった。原因は Go 1.26 以降の js/wasm 標準ライブラリがこの import を使う一方、`syumai/workers` v0.32.0 の TinyGo glue が提供しないことだった。`syumai/workers-go` v0.36.0 の生成 glue には実装が含まれるため、このモジュールへ移行して解決する。

## Decision

Worker は Go 1.27.1 と `GOEXPERIMENT=nojsonv2` でビルドし、Workers glue を `github.com/syumai/workers-go` v0.36.0 から生成する。go directive は `1.25.8` のまま維持する。(2026-09-25 追記: #8013 段階 3 で go directive を `1.27.0` に上げた。Worker は既に Go 1.27 でビルドしているため、この部分の制約は解消した。)

## Consequences

- サーバー/CLI と Worker で JSON 実装が異なり、エラー文言には差が生じる可能性がある。
- `nojsonv2` が削除される Go では再び上限超過となるため、削除前に json 依存の削減か Worker 追加が必要。
- サイズ検査だけでは wasm の import 不足を検出できないので、下記の import 検査を CI に入れた。
- Go を上げるたびに全 Worker のサイズと import を検証する。
