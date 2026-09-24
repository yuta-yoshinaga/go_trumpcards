# ADR-0042: Worker の Go 1.27 更新と encoding/json v1 の維持

## Status

Accepted

## Date

2026-09-24

## Context

Go 1.27 が公開され、Go 1.25 はセキュリティ修正の対象外になった。Worker は TinyGo 0.42.0 でビルドしている。ADR-0027 と ADR-0041 では、TinyGo が受け付ける Go の中で最小の Go 1.25 を Worker のビルドに使う前提としていた。

Go 1.27 の `encoding/json` は v2 実装になり、Makefile と同じフラグで計測すると全 Worker が gzip 1,048,576 B の上限を超える。`GOEXPERIMENT=nojsonv2` を指定すると全て上限内に収まる。以下は 2026-09-24 ローカル実測値（gzip バイト）。

| Worker | Go 1.25.8 | Go 1.27.1 | Go 1.27.1 + nojsonv2 |
|--------|----------:|----------:|---------------------:|
| casino | 938404 | 1079234 | 956531 |
| classic | 935348 | 1075400 | 953305 |
| solo | 936416 | 1076888 | 954482 |
| extra | 939691 | 1080811 | 957650 |
| extra2 | 929381 | 1070092 | 947553 |
| extra3 | 937250 | 1076482 | 954893 |
| extra4 | 928692 | 1066872 | 946997 |
| extra5 | 939211 | 1079620 | 957559 |
| extra6 | 958719 | 1098990 | 976767 |
| extra7 | 957852 | 1098448 | 975983 |

### 検討した代替案

1. **Go 1.25 に据え置く** — セキュリティ修正対象外の Go を使い続けることになるため却下する。
2. **json v2 のまま Worker を増やして再バケットする** — 容量超過には対応できるが、ADR-0041 級の Worker 追加・再バケット作業が必要となるため、まず v1 実装を維持する。

## Decision

Worker は Go 1.27.1 と `GOEXPERIMENT=nojsonv2` でビルドする。`GOEXPERIMENT` は Worker のビルドにのみ設定し、サーバーと CLI は Go 1.27 の json v2 実装を使う。

`go.mod` の `go 1.25.8` は変更しない。段階 3 の go ディレクティブ更新は、必要な言語機能が出るまで見送る。

## Consequences

- サーバーは json v2、Worker は json v1 を使うため実装が分かれる。Marshal/Unmarshal の挙動は同じだが、エラー文言には差が生じる可能性がある。
- `nojsonv2` が削除される Go では Worker が約 140 KB 増えて上限超過に戻る。それまでに json 依存を削減するか Worker を追加する必要がある。
- Go を上げるたびに `measure.sh` で 10 Worker 全てのサイズを測る。
