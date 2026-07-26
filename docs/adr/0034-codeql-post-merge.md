# ADR-0034: CodeQL を PR ではなくマージ後に実行する

## Status

Accepted

## Date

2026-07-26

## Context

CI の待ち時間を短縮する作業（PR #4343–#4351）で CI ワークフローのクリティカルパスを 1044 秒から 230 秒まで下げたが、PR が起動する**全ワークフロー**を計測したところ、待ち時間を決めていたのは CI ではなく CodeQL だった。

| ワークフロー | PR 上で最も遅いジョブ |
|---|---|
| **CodeQL — Analyze (go, autobuild)** | **459 秒** |
| CI — E2E シャード | 230 秒 |
| Cloudflare workers build (tinygo ×4) | 225 秒 |
| CodeQL — Analyze (javascript-typescript) | 218 秒 |

Go ジョブの 459 秒の内訳は、抽出 250 秒 + データベース書き出し 120 秒 + クエリ 51 秒。Go モジュールは既に `actions/setup-go` のキャッシュに乗っており（ログに `go: downloading` が 0 行）、ジョブ内部に削れる無駄はない。抽出時間は 219 ゲーム分のコンパイルそのものである。

### 検討した代替案

1. **`build-mode: none` に切り替える**（JavaScript 側は既にこれを使用）— 最も分かりやすい高速化だが、CodeQL 自身のヘルプテキストが実行ログ内で次のように述べている: *"buildless extraction will generally yield **less accurate analysis results**, and should only be used in cases where it is not possible to build the code"*。本プロジェクトはビルド可能であり、セキュリティスキャンの検出精度を速度と引き換えにするのは筋が悪い。**却下。**
2. **ジョブ内部の最適化** — キャッシュは既に効いており、抽出はコンパイル時間に等しい。**打つ手なし。**
3. **実行タイミングを変える**（採用）— スキャン内容は一切変えず、いつ走らせるかだけを変える。

## Decision

CodeQL の `pull_request` トリガーを廃止し、`push: [develop]` と週次 `schedule` のみとする。

`develop` は統合ブランチであり、リリースを駆動するのは `master`（ADR-0019 の auto-tag）である。したがってマージ直後にスキャンしても、検出はリリースよりはるかに早い。週次実行は、コードが変わらないままクエリパックが更新された場合をカバーし続ける。

ADR-0019 の「CodeQL: push/PR時にセキュリティスキャン」のうち、**PR 時の実行のみを本 ADR が置き換える**。ADR-0019 の他の決定（golangci-lint、CI、auto-tag）は有効。

## Consequences

**得るもの:**

- PR の待ち時間が約 459 秒から約 230 秒（E2E シャード）に短縮される。
- 同時実行ジョブ数が減る。`develop` への push では 25 ジョブに達し、GitHub Free の 20 ジョブ上限でシャードが待たされていた。

**失うもの（明示的なトレードオフ）:**

- **マージ前の検出シグナルが無くなる。** PR が持ち込んだ脆弱性は、マージ前ではなくマージ数分後に報告される。所見は従来どおり Security タブに現れるが、PR のチェック欄には出ない。
- `develop` はブランチ保護されておらず、CodeQL は必須チェックでもなかったため、実運用上ゲートとして機能していなかった。とはいえこれは実際の後退であり、無償の改善ではない。

**元に戻す方法:** `.github/workflows/codeql-analysis.yml` の `on:` に以下を戻すだけ（3 行）。

```yaml
  pull_request:
    branches: [ develop ]
```

マージ前スキャンを必須にしたい場合は、併せて `develop` のブランチ保護で CodeQL を required check に指定する必要がある（現在は未保護）。
