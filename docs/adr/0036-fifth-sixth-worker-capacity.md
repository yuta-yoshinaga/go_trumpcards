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
  → 実施結果は「[Phase 2 の結果](#phase-2-の結果2026-07-28-追記)」。casino だけはこの目標を**達成できていない**（理由は制約 2）。

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

## Phase 2 の結果（2026-07-28 追記）

41 ゲームを `extra2`（22）/ `extra3`（19）へ移した。実測（同一ツリーで 6 worker 一括ビルド、Go 1.25.8 + TinyGo 0.40.1 + `wasm-opt -Oz`, gzip）:

| Worker | ゲーム数 | before | after | 上限比 | 余裕 |
|--------|---------|--------|-------|--------|------|
| casino | 64 → 58 | 1,032,505 | 960,891 | 91.6% | 16.1 → **85.6 KB** |
| classic | 57 → 41 | 1,037,964 | 851,499 | 81.2% | 10.6 → **192.4 KB** |
| solo | 52 → 45 | 980,329 | 857,016 | 81.7% | 65.1 → **187.1 KB** |
| extra | 46 → 34 | 1,029,817 | 871,553 | 83.1% | 18.8 → **172.9 KB** |
| extra2 | 0 → 22 | — | 627,732 | 59.9% | **411.0 KB** |
| extra3 | 0 → 19 | — | 651,547 | 62.1% | **387.7 KB** |

**合計余裕 110 KB → 約 1,437 KB。** 新規ゲーム候補 45 件（概算 633 KB）は extra2 + extra3 の余裕（799 KB）だけで収容できる。

移動の単位・順序を決める過程で、**バケットは自由に分割できない**ことが分かった。ここが本 ADR で最も価値のある発見なので記録する。

### 制約 1: ゲームと実装は 1 対 1 ではない

6 つの実装を 16 ゲームが共有している。`razz` と `sevencardstud` は `SevenCardStud`、`spanish21` は `BlackJack`、`irishpoker` は `Pineapple` に相乗りしている（ほか `Omaha` に 4、`VideoPoker` に 3、`FreeCell` に 2）。**バケットはゲームではなく実装の属性**であり、片方だけ動かすと共有ファイルの build タグを書き換えて、残った側が「自分のバイナリに存在しないゲーム」として登録された状態になる。`move-game.py` はグループの部分移動を拒否する。

### 制約 2: バケット内が共有シンボルで溶接されている

より重い制約。`GameResult` / `GameResultWin` / `GameResultLose` / `GameResultDraw` は **`BlackJack.go` の中で宣言されている**のに casino の 19 ゲームが使う。同様に `compareHighCardsSlice` は `HoldemPlayer.go` にあって 12 ゲームが、`HoldemPlayStyle` 系は `HoldemConfig.go` にあって 6 ゲームが使う。blackjack / poker / holdem を casino から出すと、**最大 82 個のパッケージレベルシンボル**が宙に浮き、残りが軽量化どころかコンパイルすら通らない。

この 82 は**上限**として読むこと。参照は字句一致で数えているので、パッケージレベルのシンボルと同名の構造体フィールドがどこかにあると、呼んでもいないゲームが利用者として数えられる（`dealerIdx` は `BlackJackBasicStrategy.go` の関数であると同時に十数ゲームのフィールド名でもあり、実際の呼び出しは 0）。逆に**タグの無いファイルの宣言は数に入れない** — 例えば `PokerHand*` は `poker_hand_rank.go`（build タグ無し）にあり全 worker に入るので、`poker` がどこへ動いても宙に浮かない。最終的な判定は常にコンパイラ。

自由に動かせる（他ユニットと一切シンボルを交換しない）ユニット数は次のとおり:

| Worker | 自由に動かせるユニット | 動かせるソース量 |
|--------|----------------------|-----------------|
| casino | 6 / 55 | 329 KB |
| classic | 16 / 57 | 784 KB |
| solo | 23 / 51 | 1,138 KB |
| extra | 16 / 46 | 959 KB |

**casino だけは目標の余裕を再バケットでは達成できない。** 必要な 168 KB に対し、動かせるのは約 66 KB 分しかない。casino を解放するには `GameResult` / `PokerHand*` などの共有基盤をゲームファイルから共有ファイルへ切り出す必要があり、これは「サイズのための再バケット」とは別種の変更なので**本 ADR では実施せず、[#4457](https://github.com/yuta-yoshinaga/go_trumpcards/issues/4457) に切り出した**。casino が 1 MB に触れた時点で先送りはできなくなる。

なお、この 2 つの制約は `crossrefs.py` が両方向＋共有ヘルパーの 3 観点で静的に報告する。以前は TinyGo ビルドを 1 回 3.5 分かけて `undefined: X` を 1 バッチずつ潰していた作業が、1 秒で済む。

### Phase 1 に空いていた穴

Phase 1 は「ビルドとデプロイ経路が通ること」を確認したが、**バケットを増やしても壊れないことは確認していなかった**。Phase 2 で 3 つ見つかった:

- `AllCategories()` が新バケットを含んでおらず、`trumpcards games` の一覧から 41 ゲームが黙って消えた。この関数の doc コメントは「Category を足したらこのスライスも足す必要がある、それが SSoT の保証」と書いてあったが、**コメントは保証にならなかった**。
- `TestCategoryString` / `TestAllEntriesAreValid` が 4 カテゴリ決め打ちだった。

いずれも「書いておいたルールが守られなかった」型なので、テストで機械的に縛った:

- `TestAllCategoriesCoversEveryDeclaredCategory` — 宣言済み Category を走査し、`AllCategories()` の取りこぼしを落とす
- `TestDocsMatchRegistry` — `docs/cloudflare-workers.md` の worker 別一覧をレジストリと突き合わせる。この表は散文で手書きされていたため 4 箇所が古いまま残っており（4 つ目の worker 追加後も「3 つ」と書かれていた）、生成された平坦なキー一覧に置き換えた

### マージ順序（重要）

`deploy-cloudflare.yml` は `internal/**` と `frontend/**` の push で走る。Phase 2 は必ず両方に触るため、**Cloudflare 側の準備が済む前にマージすると 41 ゲームが本番で 404 になる**（既存 4 worker が移動後のゲームを含まない状態で再デプロイされ、extra2/extra3 は matrix に無く、フロントは空 URL へルーティングする）。Phase 1 と違い「新 worker が増えないだけ」では済まない。順序は次のとおり:

1. KV namespace を作成し `wrangler.toml` の `REPLACE_ME_*` を差し替える
2. `WORKER_EXTRA{2,3}_URL` / `_STAGING_URL` と `VITE_WORKER_EXTRA{2,3}_URL` を設定する
3. deploy matrix に `extra2` / `extra3` を追加する（1 と同じ変更で）
4. **その後に** Phase 2 をマージする

## Consequences

### 良い点

- 新規ゲーム 45 件を収容でき、classic / solo / extra には 170 KB 以上の余裕が戻る（実測値は Phase 2 の表を参照）。**casino だけは 86 KB 止まり**で、理由と対処は上の「制約 2」に書いた。
- ローカルでサイズを実測できるようになったため、再バケットの試行が CI 一周（約 20 分）から数分に短縮される。
- Worker 追加手順が記録され、7 つ目が必要になったときに再発見しなくて済む。

### 悪い点・受け入れるリスク

- Worker が 6 つになり、デプロイ対象・KV namespace・環境変数がそれぞれ 6 系統になる。運用面の複雑さは単調に増える。
- Worker を 1 つ増やすたびに **232 KB のベースラインを丸ごと払う**（6 つで約 1.4 MB がランタイムの重複）。分割で容量を稼ぐやり方は本質的に効率が悪く、増やすほど悪化する。
- 無料枠の Worker 数上限（100）には余裕があるが、1 プロジェクトで 6 つ使うのは多い。**7 つ目が必要になった時点で、分割の継続ではなく有料プラン（上限 10 MB）への移行を再検討すべき**。本 ADR はその判断を先送りしているに過ぎない。
- `Category` の値が増えるほど「どのバケットに入れるべきか」の判断がゲーム追加時のコストになる。緩和策として、新規ゲームは常に**最も余裕のある Worker** に入れる運用とし、意味的な分類は一切考えない。
- **casino の共有基盤問題を先送りした。** `GameResult` が `BlackJack.go` にあるような配置は、そのファイルを含むゲームをバケットの人質にする。casino は今回 86 KB の余裕しか得られておらず、次に容量が問題になるのはここである。切り出しを避け続けるほど、選択肢は「casino を丸ごと 1 worker にする」か「有料プラン」に狭まる。

## References

- [ADR-0027](0027-cloudflare-workers-wasm.md) Worker 分割の原型
- [ADR-0032](0032-fourth-worker-capacity.md) 4 つ目の追加。ゲームを移す手順はこちら
- [#4375](https://github.com/yuta-yoshinaga/go_trumpcards/issues/4375)–[#4419](https://github.com/yuta-yoshinaga/go_trumpcards/issues/4419) 収容先を必要としている新規ゲーム候補 45 件
- [docs/cloudflare-workers.md](../cloudflare-workers.md) Worker ごとのゲーム一覧とビルド手順
- [#4457](https://github.com/yuta-yoshinaga/go_trumpcards/issues/4457) casino の共有基盤切り出し（制約 2 の対処）
