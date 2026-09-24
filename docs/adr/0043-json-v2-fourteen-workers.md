# ADR-0043: Worker を 14 個に増やし encoding/json v2 に戻す

## Status

Accepted

## Date

2026-09-25

## Context

ADR-0042 は Worker ビルドにだけ `GOEXPERIMENT=nojsonv2` を付けて json v1 に戻した。nojsonv2 は将来の Go で削除予定で、サーバ（v2）と Worker（v1）で JSON 実装が分かれていた。json v2 のまま 10 Worker をビルドした実測（TinyGo 0.42.0 / Go 1.27.1 / wasm-opt -Oz、gzip）は次の通りで、全て 1,048,576 B を超過した（約 122 KB/Worker）。

| Worker | gzip bytes |
|--------|------------:|
| casino | 1,079,202 |
| classic | 1,075,368 |
| solo | 1,076,859 |
| extra | 1,080,756 |
| extra2 | 1,069,951 |
| extra3 | 1,076,339 |
| extra4 | 1,066,649 |
| extra5 | 1,079,547 |
| extra6 | 1,098,903 |
| extra7 | 1,098,402 |

空 Worker は v2 で 342,756 B であり、増分は固定費ではなくゲームコード量に比例する（約 19%）。

検討した代替案は、(a) nojsonv2 を維持する（削除時に再び上限を超える）、(b) Worker 内の JSON 利用を減らす（1,675 個の MarshalJSON に手を入れる必要がある）、(c) Worker を増やす、の3つ。(c) を採用した。3 個増では平均約 1,013 KB で余裕がなく、4 個増（extra8〜extra11）とした。

## Decision

Worker を 14 個にし（extra8、extra9、extra10、extra11）、ゲーム 99 個を再バケットして `GOEXPERIMENT=nojsonv2` を撤去する。Worker とサーバの両方で `encoding/json` v2 を使う。

### 改訂 (2026-09-25)

マージ後の staging で全 383 ゲームの reset を実行したところ、gostop / basra / koikoi の Worker が停止した。原因は TinyGo 0.42.0 の reflect が `reflect.SliceOf` / `reflect.MapOf` を実装しておらず (`panic: unimplemented: reflect.SliceOf()`)、Go 1.27 の encoding/json v2 がこれらを呼ぶことだった。v2 は文字列以外のキーを持つ map を 2 要素以上マーシャルするとき、決定的出力のためキーをソートする経路で `src/encoding/json/v2/arshal_default.go:910-911` の `SliceOf` を呼び、空でない map のアンマーシャルでは同ファイル 1012 行の `MapOf` を呼ぶ。最小再現は TinyGo wasm で `map[int][]int{0: {5}, 3: {1, 2}}` を `json.Marshal` すると panic し、1 要素なら通る。

Worker 14 個と再バケットは維持し、`GOEXPERIMENT=nojsonv2` を戻す。json v2 へ移れる条件は、(a) TinyGo が `reflect.SliceOf` / `reflect.MapOf` を実装する、または (b) Worker が JSON にする型から文字列以外のキーの map を無くす、のどちらか。Web 出力の `map[int]...` だけで20箇所以上あり、KV に保存するセッション状態も対象となる。移行時は staging で全ゲームを reset するだけでなく、数手進めて確かめる。サイズ・import 検査・Go のテストはいずれもこの停止を検出しない。

再バケット後の実測（json v2、gzip B / 残り KB）は次の通り。

| Worker | gzip B | 残り KB |
|--------|-------:|--------:|
| casino | 968,939 | 77.8 |
| classic | 949,271 | 97.0 |
| solo | 904,013 | 141.2 |
| extra | 955,307 | 91.1 |
| extra2 | 959,624 | 86.9 |
| extra3 | 964,109 | 82.5 |
| extra4 | 949,308 | 96.9 |
| extra5 | 970,603 | 76.1 |
| extra6 | 956,766 | 89.7 |
| extra7 | 961,299 | 85.2 |
| extra8 | 957,897 | 88.6 |
| extra9 | 974,168 | 72.7 |
| extra10 | 872,526 | 171.9 |
| extra11 | 882,478 | 162.2 |

最小の余裕は extra9 の 72.7 KB、合計余裕は約 1,420 KB。新規ゲームは extra10 / extra11 から入れる。

移動時、casino のテーブルゲーム（blackjack 系・video poker 系等）は共有ファイルに `|| extraN` タグを足せば移せることが分かった。一方、casinoholdem / ultimatetexasholdem / texasholdembonus は Holdem、threecard は chinesepoker の presenter、indianrummy / pan / fivehundred（後に euchre ごと移動）/ bolivia は他ゲームの実装に依存するため残した。

## Consequences

- KV namespace を12個、repo variable を8個追加した。
- Worker は14個になり、デプロイのジョブも14個になった。
- サーバと Worker の JSON 実装が揃う。
- Cloudflare 無料枠の Worker 数上限（100）には余裕がある。
- CI の Worker ビルド workflow が go.mod / go.sum の変更でも走るようにした（#8048 で走らなかった穴）。
- `nojsonv2` が Go から削除されるまでに上の条件のどちらかが要る。サイズ面の準備（14 Worker）は済んでいる。
