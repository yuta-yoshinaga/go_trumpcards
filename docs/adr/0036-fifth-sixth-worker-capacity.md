# ADR-0036: 5 つ目・6 つ目の Cloudflare Worker（容量バケット）の追加

## Status

Accepted

## Date

2026-07-28

## Context

[ADR-0032](0032-fourth-worker-capacity.md) で 4 つ目の Worker (`extra`) を追加してから 219 ゲームまで増え、**4 Worker すべてが再び上限間際**になった。

ローカル実測（2026-07-28, CI と同一の Go 1.25 + TinyGo 0.40.1 + `wasm-opt -Oz`, gzip）:

| Worker | ゲーム数 | gzip | 上限比 | 余裕 |
|--------|---------|------|--------|------|
| casino | 64 | 1,032,505 | **98.5%** | 16.1 KB |
| classic | 57 | 1,037,964 | **99.0%** | 10.6 KB |
| solo | 52 | 980,329 | 93.5% | 65.1 KB |
| extra | 46 | 1,029,817 | **98.2%** | 18.8 KB |

**合計余裕 約 110 KB。**

1 ゲームあたりのコストは、空の Worker を実際にビルドして分離した。**中身ゼロの Worker が 232 KB**（Go ランタイム + syumai/workers + 共有ヘルパー）で、これは games 数によらず固定でかかる。

| Worker | ゲーム数 | gzip | − ベースライン | 1 ゲームあたり |
|--------|---------|------|---------------|---------------|
| casino | 64 | 1,032,505 | 800,489 | 12,508 B |
| classic | 57 | 1,037,964 | 805,948 | 14,139 B |
| solo | 52 | 980,329 | 748,313 | 14,391 B |
| extra | 46 | 1,029,817 | 797,801 | 17,344 B |

**限界コストは平均 14.4 KB/ゲーム。** 「総 gzip ÷ 総ゲーム数」で 18.6 KB と見積もると、固定の 232 KB をゲームに按分してしまい 3 割ほど過大になる。この差は「あと何ゲーム入るか」の判断に直接効くので、ベースラインは必ず分離して考えること。

したがって 1 Worker に載るゲームは **約 57**（(1,048,576 − 232,016) ÷ 14,395）。**4 Worker 合計であと 6 ゲーム程度**しか入らない。

新規ゲーム候補 45 件（[#4375](https://github.com/yuta-yoshinaga/go_trumpcards/issues/4375)–[#4419](https://github.com/yuta-yoshinaga/go_trumpcards/issues/4419)）の総量は概算 **633 KB**。**容量が新規ゲーム追加のブロッカー**という ADR-0032 と同じ状況が再来した。

なお ADR-0032 は「TinyGo のバイナリサイズはローカルで計測できないため CI のレポートを見ながら調整する」としていたが、**この制約は解消した**。TinyGo 0.40.1 と Go 1.25.8 をローカルに導入済みで、手順は [reference: ローカル Worker ビルド](../cloudflare-workers.md#ローカルでサイズを実測する) にある。再バケットのたびに CI を一周させる必要はなくなった。

### 検討した代替案

1. **1 つだけ追加して詰め込む** — 新 Worker 1 つが持てるのは約 57 ゲーム分（816 KB）。45 件（633 KB）は入るが使い切りに近く、既存 4 つの余裕（110 KB）は 1 バイトも回復しない。追加直後にまた上限問題に戻る。
2. **有料プランに移行して Worker 分割自体をやめる** — 上限が 10 MB になり分割が不要になる。技術的には最も単純だが、本プロジェクトは無料枠で運用する前提であり、課金判断は技術的決定の範囲外。
3. **コードサイズ削減で凌ぐ** — ADR-0032 で既に検討し「数 KB 規模で足りない」と結論済み。219 ゲーム時点でも状況は同じ。
4. **5 つ目と 6 つ目を同時に追加する** — 新規 45 件を収容したうえで、既存 4 Worker にも 150 KB 以上の余裕を戻せる。

## Decision

**案 4 を採る。** size バケットを 2 つ追加する:

| Category 定数 | Worker 名 | Cloudflare 名 |
|---------------|-----------|---------------|
| `CategoryExtra2` | `extra2` | `go-trumpcards-extra2` |
| `CategoryExtra3` | `extra3` | `go-trumpcards-extra3` |

`Category` が **ユーザー向け分類ではなく純粋なサイズバケット**である原則は ADR-0027 / ADR-0032 から変わらない。`extra2` / `extra3` という名前は、その原則を名前自体で示すために意図的に無味乾燥にしてある（`puzzle` のような意味ありげな名前は、次に誰かが「このゲームは分類的にここだろう」と考え始める余地を作る）。

### 段階を分ける

ADR-0032 と同じく 2 フェーズで進める。1 フェーズ目で空のバケットを足して CI が通ることを確かめ、2 フェーズ目で中身を動かす。

- **Phase 1（本 ADR と同時）**: 空の `extra2` / `extra3` を作る。ビルド・サイズチェック・デプロイ経路が通ることだけを確認する。ゲームは 1 つも動かさない。
- **Phase 2**: 既存 4 Worker から再バケットし、各 Worker に 150 KB 以上の余裕を作る。ローカル実測で 1 バッチずつ確認する。

### Worker を 1 つ増やすときに触る箇所

ADR-0032 は「ゲームを移す手順」しか残していなかったため、ここに Worker 追加側の手順を記録する。

**コード（11 箇所）**

1. `cmd/workers/<name>/main.go` — エントリポイント
2. `internal/infrastructure/games/<name>/<name>.go` — カテゴリ sub-package
3. `workers/<name>/wrangler.toml` — KV namespace ID が必要
4. `internal/infrastructure/games/registry.go` — `Category<Name>` 定数
5. `internal/infrastructure/games/registry_test.go` — カテゴリ別カウント
6. `Makefile` — `build-worker-<name>` と `build-workers` への追加
7. `.github/workflows/cloudflare-workers-build.yml` — matrix
8. `.github/workflows/deploy-cloudflare.yml` — URL 変数。**matrix への追加は KV ID を
   実物に差し替える変更と同時に行うこと**（下記）
9. `frontend/src/api/gameExec.ts` — `WORKER_<NAME>` 定数
10. `frontend/vite-env.d.ts` — `VITE_WORKER_<NAME>_URL`
11. `docs/cloudflare-workers.md` — 一覧

**Cloudflare アカウント側（コードでは完結しない）**

- KV namespace を production / preview / staging 分作成し、その ID を `wrangler.toml` に入れる
- GitHub の deploy variables に `WORKER_<NAME>_URL` と `WORKER_<NAME>_STAGING_URL` を設定
- フロントエンドビルド用に `VITE_WORKER_<NAME>_URL` を設定

**この 3 つが揃うまで新 Worker は実際にはデプロイできない。** Phase 1 の PR は KV ID をプレースホルダで置き、ビルドとサイズチェックまでを通す。

**プレースホルダのまま deploy matrix に足してはいけない。** `deploy-cloudflare.yml` は
`cmd/workers/**` と `workers/**` の変更で develop への push 時に走るため、Phase 1 の変更
そのものがトリガーになる。プレースホルダ ID で `wrangler deploy` が失敗し、`fail-fast` は
既定 true なので**他の 4 Worker のデプロイまで巻き添えでキャンセル**され、さらに
`deploy-pages` が `needs: deploy-workers` なのでフロントエンドのデプロイも止まる。
つまり「新 Worker がデプロイされないだけ」では済まず、**デプロイ経路全体が壊れる**。
build 専用の `cloudflare-workers-build.yml` は wrangler も KV も触らないので Phase 1 で
足して問題ない。Makefile の `deploy-workers` ターゲットも同じ理由で 4 つのままにしてある。

## Consequences

### 良い点

- 新規ゲーム 45 件を収容でき、かつ既存 4 Worker に 150 KB 以上の余裕が戻る。6 Worker 体制の総容量は約 340 ゲーム分で、219 ゲームの現状に対し **約 120 ゲーム分の余地**ができる。
- ローカルでサイズを実測できるようになったため、再バケットの試行が CI 一周（約 20 分）から数分に短縮される。
- Worker 追加手順が記録され、7 つ目が必要になったときに再発見しなくて済む。

### 悪い点・受け入れるリスク

- Worker が 6 つになり、デプロイ対象・KV namespace・環境変数がそれぞれ 6 系統になる。運用面の複雑さは単調に増える。
- Worker を 1 つ増やすたびに **232 KB のベースラインを丸ごと払う**（6 つで約 1.4 MB がランタイムの重複）。分割で容量を稼ぐやり方は本質的に効率が悪く、増やすほど悪化する。
- 無料枠の Worker 数上限（100）には余裕があるが、1 プロジェクトで 6 つ使うのは多い。**7 つ目が必要になった時点で、分割の継続ではなく有料プラン（上限 10 MB）への移行を再検討すべき**。本 ADR はその判断を先送りしているに過ぎない。
- `Category` の値が増えるほど「どのバケットに入れるべきか」の判断がゲーム追加時のコストになる。緩和策として、新規ゲームは常に**最も余裕のある Worker** に入れる運用とし、意味的な分類は一切考えない。

## References

- [ADR-0027](0027-cloudflare-workers-wasm.md) Worker 分割の原型
- [ADR-0032](0032-fourth-worker-capacity.md) 4 つ目の追加。ゲームを移す手順はこちら
- [#4375](https://github.com/yuta-yoshinaga/go_trumpcards/issues/4375)–[#4419](https://github.com/yuta-yoshinaga/go_trumpcards/issues/4419) 収容先を必要としている新規ゲーム候補 45 件
- [docs/cloudflare-workers.md](../cloudflare-workers.md) Worker ごとのゲーム一覧とビルド手順
