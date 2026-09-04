# ADR-0038: 8 つ目の Cloudflare Worker（容量バケット）の追加

## Status

Accepted

## Date

2026-09-05

## Context

[ADR-0037](0037-seventh-worker-capacity.md) で 7 つ目の Worker (`extra4`) を
追加してから 368 ゲームまで増え、**7 Worker すべてが再び上限間際（上限 1,048,576 B gzip まで 35 KB 以内）**になった。

起点の issue は [#7073](https://github.com/yuta-yoshinaga/go_trumpcards/issues/7073)（測定の記録）。[#7103](https://github.com/yuta-yoshinaga/go_trumpcards/issues/7103) が実装 issue である。

ローカル実測（2026-09-05, CI と同一の Go 1.25.8 + TinyGo 0.40.1 + `wasm-opt -Oz`, gzip。CI と一致することを casino で確認: CI 1,012,988 / ローカル 1,012,998）。
無料枠の上限は **1,048,576 B (1 MB gzip)**:

| Worker | ゲーム数 | gzip | 上限比 | 余裕 |
|--------|---------|------|--------|------|
| extra4 | 50 | 1,033,804 | 98.6% | 14.4 KB |
| extra | 44 | 1,030,681 | 98.3% | 17.5 KB |
| extra2 | 55 | 1,021,116 | 97.4% | 26.8 KB |
| classic | 53 | 1,017,242 | 97.0% | 30.6 KB |
| extra3 | 44 | 1,016,117 | 96.9% | 31.7 KB |
| solo | 56 | 1,013,051 | 96.6% | 34.7 KB |
| casino | 66 | 1,012,998 | 96.6% | 34.7 KB |

**合計余裕 190.4 KB。**

#7073 が挙げた選択肢 B（タグ無し domain にビルドタグを付けて他 Worker から外す）は、**PR #7086 でハーツ 1 ゲームをパイロットして否定済み**である。1 ゲームあたり 129〜252 バイト、extra4 は +49 バイトしか減らなかった。TinyGo の DCE（Dead Code Elimination）が既に不要コードを落としているためである。368 ゲーム全部に適用しても 8 個目の Worker 1 つ分（233 KB）に届かない。

新規ゲーム候補 15 件（[#7088](https://github.com/yuta-yoshinaga/go_trumpcards/issues/7088)–[#7102](https://github.com/yuta-yoshinaga/go_trumpcards/issues/7102)）が起票済みであり、現在の合計余裕 190.4 KB では到底入り切らない。**容量が新規ゲーム追加のブロッカー**という過去の ADR と同じ状況が四度目になった。

中身ゼロの Worker（固定コスト）について、空の `extra5` を実際にビルドして確認したところ、**234,036 B gzip / 余裕 795.4 KB** であった。ADR-0036 / ADR-0037 が記録した「中身ゼロで約 233 KB」を再現している（extra4 は 233,341 B、差 +695 B）。

### 検討した代替案

1. **既存 7 つの間で詰め替えるだけ** — 合計余裕は 190.4 KB のまま増えないのでブロッカーは解けない。
2. **有料プランへ移行して分割をやめる** — 上限が 10 MB になり分割が不要になる。課金判断は技術的決定の範囲外（ADR-0036 / ADR-0037 と同じ理由）。
3. **コードサイズ削減で凌ぐ** — #7073 のパイロット（PR #7086）の実測により否定された（上記）。過去の ADR では概算により見送っていたが、**今回は初めて実測による裏取りがある**点が異なる。
4. **8 つ目を足し、同時に既存 7 つから再バケットする** — 空の Worker が約 795 KB の余裕を持つため、15 件の新規ゲームを吸収した上で既存 Worker の余裕を回復できる。

## Decision

**案 4 を採る。** size バケットを 1 つ追加する:

| Category 定数 | Worker 名 | Cloudflare 名 |
|---------------|-----------|---------------|
| `CategoryExtra5` | `extra5` | `go-trumpcards-extra5` |

`Category` が **ユーザー向け分類ではなく純粋なサイズバケット**である原則は ADR-0027 / ADR-0032 / ADR-0036 / ADR-0037 から不変である。`extra5` という無味乾燥な名前も同じ理由で意図的に選ばれている。

### フェーズを分けない

ADR-0037 と同様、**1 つの PR でバケット追加と再バケットを同時に行う**。

ただし理由は ADR-0037 と異なる。ADR-0037 では `extra3` の残り容量が 188 バイトという極度の危険区間にあったが、今回は最少でも `extra4` の 14.4 KB であり、そのような危険区間はない。分けない理由は、**空バケットを足すだけでは他 Worker が 1 バイトも回復せず、#7088–#7102 のブロッカーが解けないから**である。

なお、Cloudflare 側の実体はコミット前に用意した（ADR-0036 が記録した「プレースホルダのまま deploy matrix に足さない」制約に従う）:
- KV namespace 3 つ（production / preview / staging を別々に作成）
- repo variables `WORKER_EXTRA5_URL` / `WORKER_EXTRA5_STAGING_URL`

### 移動対象の選び方（使ったツールと、その限界）

移せるユニット（移動しても取り残しシンボルが出ない組）の抽出には、リポジトリの
`crossrefs.py` をそのままは使っていない。**この再バケットでは、全ファイルを 1 回だけ解析して
全ユニットを一括判定する使い捨てスクリプトを別に書いて選定した。リポジトリには入れていない。**
理由と、その過程で分かったことを残す。

1. **`crossrefs.py` は 1 ユニットにつき全ソースを読み直す。** 実測で 1 回 120 秒を超え、
   352 ユニットに回すと 12 時間規模になる。1 件を詳しく見るには適しているが、
   「どれなら移せるか」を全件について知る用途には向かない。
2. **ゲーム名からファイルを探す方式は数え落とす。** `crossrefs.py` はゲームの Go 型名で
   ファイルを探すので、`internal/domain/interfaces/` の snake_case ファイルや、複数バケットが
   共有するタグ付きヘルパを拾えない（ADR-0037 が「5 件の見落とし」として記録したのと同じ原因）。
   ビルドタグを直接読む形で数え直すと、同じ casino ユニットについて「残る側が失うシンボル」
   11 件は `crossrefs.py` と一致し、「移る側が欠くシンボル」は 28 → 66 件になった。
   **後者の方が厳しく、コンパイラの結果と整合していた。**
3. **ファイルの共有だけでユニットを作ると、シンボルで溶接された組を取りこぼす。**
   `kingo` と `oichokabu` は単独ではどちらも移せない（`kingo` が `buildOichoKabuDeck` など
   3 シンボルを `oichokabu` に依存している）が、**2 つ一緒なら clean**。この 2 つはファイルを
   共有していないので、ファイル単位のユニット分割では 1 つにならなかった。
   加えて `speculation` は、ブランク識別子 `_` を実シンボルとして数えていたスクリプト側の
   バグで候補から外れていた。**この 2 件を拾えたことで `extra` が目標に届いた**
   （拾う前は移せるユニットを使い切っても 84.5 KB で頭打ちだった）。

**次に再バケットする人へ**: この一括判定スクリプトをリポジトリの
`.claude/skills/rebucket-game/scripts/` に入れる価値はあるが、本 ADR の変更範囲外とした。

移動後のバケット境界の破綻検証には `GOOS=js GOARCH=wasm go build -tags <worker> -o /dev/null ./cmd/workers/<worker>`
を用い、2 ラウンドとも 8 Worker すべてで 0 件（破綻なし）を確認した。

### `casino` は目標に届かない（制約の悪化）

各 Worker に 100 KB 以上の余裕を持たせることを目標としたが、**`casino` だけは達成できない**（ADR-0037 の制約が悪化）。

ADR-0037 の時点では「動かせるのは 8 コンポーネント / 364 KB」だったが、今回の解析では **clean なユニットが 0** であった。`GameResult` が `BlackJack.go` に、`compareHighCardsSlice` が `HoldemPlayer.go` にあるという同一の溶接が原因である。

これにより casino は 34.7 KB の余裕のまま動かない。これを解くには共有シンボルをタグ無しファイルへ切り出すリファクタが必要となるが、これは**本 ADR の範囲外**とし、casino が上限を割った時点で改めて判断する。

### `TestDocsMatchRegistry` の検査漏れの解消

テスト整備の過程で `TestDocsMatchRegistry` の穴を発見し解消した。

従来のテストは行数だけ `len(games.AllCategories())` から導いていたが、**中身の突き合わせは `casino, classic, solo, extra, extra2, extra3` の 6 バケットがハードコード**されており、7 個目（`extra4`）と 8 個目（`extra5`）のバケットを検査していなかった。同じ関数内に「ADR-0037 で `!= 6` と書いてあって壊れた」というコメントがありながら、同じ誤りが残っていた。

これを `games.AllCategories()` の走査に変更し、**extra5 の行だけを件数を保ったまま壊す負の対照**でテストが落ちることを確認した（`only in docs: [zzz] / only in registry: [daifugo]`）。

## 実測結果

31 ゲームを `extra5` へ移した（ラウンド 1: 23 ゲーム / 1,937 KB ソース、ラウンド 2: 8 ゲーム / 544 KB ソース）。
ローカル実測（CI と同一の Go 1.25.8 + TinyGo 0.40.1 + `wasm-opt -Oz`, gzip）。上限 1,048,576 B:

| Worker | ゲーム数 | before gzip | after gzip | 上限比 | before 余裕 | **after 余裕** |
|--------|--------:|------------:|-----------:|-------:|------------:|---------------:|
| `casino` | 66 | 1,012,998 | 1,013,002 | 96.6% | 34.7 KB | **34.7 KB** |
| `classic` | 50 | 1,017,242 | 943,345 | 90.0% | 30.6 KB | **102.8 KB** |
| `solo` | 51 | 1,013,051 | 939,769 | 89.6% | 34.7 KB | **106.2 KB** |
| `extra` | 36 | 1,030,681 | 936,576 | 89.3% | 17.5 KB | **109.4 KB** |
| `extra2` | 50 | 1,021,116 | 930,675 | 88.8% | 26.8 KB | **115.1 KB** |
| `extra3` | 39 | 1,016,117 | 942,136 | 89.8% | 31.7 KB | **103.9 KB** |
| `extra4` | 45 | 1,033,804 | 935,896 | 89.3% | 14.4 KB | **110.0 KB** |
| `extra5` | 31 | — | 872,505 | 83.2% | — | **171.9 KB** |
| **合計** | **368** | | | | **190.4 KB** | **854.0 KB** |

（* 移動前の空の extra5 は 234,036 B gzip / 余裕 795.4 KB。移動前合計の 190.4 KB は既存 7 Worker の余裕の合計）

**合計余裕は 190.4 KB → 854.0 KB。**

`casino` を除く 7 Worker が目標の 100 KB を超えた。casino が 34.7 KB にとどまる理由は Decision に記したとおり、動かせる clean なユニットが 0 のためである。

### ラウンドを 2 回に分けた理由

ラウンド 1 では 23 ゲーム（1,937 KB ソース）を移動したが、**gzip/ソース比はバケットごとに 0.171〜0.229 とばらつく**ことが判明した。
ADR-0037 は 0.27〜0.61 と報告していたが、**同じ係数を使い回せないことが 2 度目に確認された**。
ラウンド 1 を 0.26 と見積もった結果、5 バケットが目標の 100 KB に届かなかった。

そのため、ラウンド 2 ではラウンド 1 で実測した比率（solo: 0.212, extra: 0.193, extra2: 0.215, extra3: 0.189, extra4: 0.171）を用いて 8 ゲーム（544 KB ソース）を選び直して組み直し、casino を除く全バケットで目標の 100 KB を達成した。

## Consequences

### 良い点

- 新規ゲーム 15 件（[#7088](https://github.com/yuta-yoshinaga/go_trumpcards/issues/7088)–[#7102](https://github.com/yuta-yoshinaga/go_trumpcards/issues/7102)）を収容可能な容量が確保された（合計余裕 854.0 KB、extra5 の余裕 171.9 KB）。
- `casino` を除く 7 Worker で 100 KB 以上の余裕が回復した。
- `TestDocsMatchRegistry` の検査漏れが修正され、7 個目・8 個目以降のバケットも自動で網羅的に検証されるようになった。
- 移動対象の選び方について、`crossrefs.py` の限界（1 ユニットごとに全ソースを読み直す、ゲーム名でファイルを探すので snake_case のファイルを数え落とす）と、ファイル単位のユニット分割がシンボル溶接組を取りこぼすことを記録できた。**ただし一括判定スクリプト自体はリポジトリに入っていない**ので、次に再バケットする人は同じものを書き直すことになる。

### 悪い点・受け入れるリスク

- **バケット移動により進行中の KV セッションが失われる。**
  各 Worker の `GAME_SESSIONS` バインディングは自身の KV namespace だけを指している（`internal/infrastructure/worker/register.go`）。移した 31 ゲームのセッションは旧 Worker の namespace に残り、到達不能になる。プレイ中だったセッションは次回リクエスト時に無効となりリセットされる。ADR-0032 / ADR-0036 / ADR-0037 でも同様の事象が発生していたが、過去の ADR には記録されていなかったため、今回初めて明記する。
- **`casino` の余裕が 34.7 KB のまま据え置かれた。**
  clean なユニットが 0 であり、再バケットのみでの解消が不可能となった。今後 casino に属するゲームの改修や追加が必要になった場合は、`BlackJack.go` や `HoldemPlayer.go` にある共有シンボルをタグ無しファイルへ切り出すリファクタリングが不可避となる。
- Worker が 8 つになり、デプロイ対象・KV namespace・環境変数の管理負荷がさらに増加した。固定コスト（約 233 KB）の重複も計約 1.87 MB に達している。

## References

- [ADR-0027](0027-cloudflare-workers-wasm.md) Cloudflare Workers (TinyGo/Wasm) によるエッジデプロイ
- [ADR-0028](0028-kv-session-persistence.md) Cloudflare KV によるセッション永続化
- [ADR-0032](0032-fourth-worker-capacity.md) 4 つ目の Cloudflare Worker（容量バケット）の追加
- [ADR-0036](0036-fifth-sixth-worker-capacity.md) 5 つ目・6 つ目の Cloudflare Worker（容量バケット）の追加
- [ADR-0037](0037-seventh-worker-capacity.md) 7 つ目の Cloudflare Worker（容量バケット）の追加
- [#7073](https://github.com/yuta-yoshinaga/go_trumpcards/issues/7073) 容量測定と選択肢の記録
- [#7086](https://github.com/yuta-yoshinaga/go_trumpcards/pull/7086) タグ無し domain のビルドタグ化パイロット（ハーツ）
- [#7088](https://github.com/yuta-yoshinaga/go_trumpcards/issues/7088)–[#7102](https://github.com/yuta-yoshinaga/go_trumpcards/issues/7102) 新規ゲーム候補 15 件
- [#7103](https://github.com/yuta-yoshinaga/go_trumpcards/issues/7103) 8 つ目の Worker 追加と 31 ゲームの再バケット
- [docs/cloudflare-workers.md](../cloudflare-workers.md) Worker ごとのゲーム一覧とビルド手順
