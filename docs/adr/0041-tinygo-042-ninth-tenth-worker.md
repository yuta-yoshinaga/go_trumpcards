# ADR-0041: TinyGo 0.42.0 への更新と 9・10 個目の Worker（容量バケット）の追加

## Status

Accepted

Worker の Go バージョンについては ADR-0042 で更新する。

## Date

2026-09-23

## Context

TinyGo を 0.40.1 から 0.42.0 に上げる。0.41.0 の tinygo-org/tinygo#5304 では method-set ベースの reflect `AssignableTo` / `Implements` が実装され、#4277 が close された。動的 `AssignableTo` が未実装で panic していたため `errors.As` が Worker を落としていた問題の原因が直った。一方、0.41.x は js/wasm で `net/http/roundtrip_js.go:73: t.roundTrip undefined` となりビルドできない。0.42.0 のリリースノートには js/wasm の roundtrip 修正があり、利用できるのは 0.42.0 だけである。0.42.0 は Go 1.25〜1.27 を受け付けるため、`go.mod`（`go 1.25.8`）と CI の Go は変更しない。

TinyGo 0.42.0 では全 Worker の gzip サイズが約 89〜103 KB 増え、5 Worker が上限 1,048,576 B を超えた（CI run 35844633939、fail-fast なし）。左が 0.40.1 の develop、右が 0.42.0:

| Worker | 0.40.1 | 0.42.0 |
|--------|-------:|-------:|
| casino | 1,022,147 | 1,125,042 |
| extra | 969,299 | 1,065,783 |
| extra5 | 960,320 | 1,060,796 |
| solo | 955,697 | 1,051,411 |
| extra2 | 960,822 | 1,050,232 |
| extra4 | 946,339 | 1,046,519 |
| classic | 947,922 | 1,045,474 |
| extra3 | 946,567 | 1,043,731 |

中身ゼロの Worker のローカル実測（Go 1.25.8 + `wasm-opt -Oz`）は 0.40.1 で 235,099 B、0.42.0 で 318,596 B。増分の大半（83.5 KB）は Worker ごとの固定費である。0.42.0 のリリースノートにある「runtime error の panic を recover 可能にした」変更（tinygo-org/tinygo#5550）が最有力候補だが、原因は切り分けておらず、これは推測である。

### 検討した代替案

1. **0.40.1 に留まる** — `TestNoErrorsAsAnywhereUnderInternal` による `errors.As` 回避策で実害はない。しかし今後の TinyGo も同じ固定費を持つ見込みで、いずれ同じ判断が必要になる。
2. **`-panic=trap` でサイズを戻す** — 全 Worker の `internal/infrastructure/recoverymw` の `recover()` が効かなくなり、panic が CORS ヘッダなしの Cloudflare 1101 になるため却下する。
3. **有料プランにする** — ADR-0036〜0038 と同じく、課金は技術判断の範囲外である。
4. **Worker を 1 つ追加して 9 つにする** — 平均余裕の見積りは約 70 KBで、新規ゲーム数件で再び詰まる。
5. **Worker を 2 つ追加して 10 にし、再バケットする** — 採用する。

## Decision

案 5 を採る。`CategoryExtra6` / `CategoryExtra7`、Worker 名 `extra6` / `extra7`、Cloudflare 名 `go-trumpcards-extra6` / `go-trumpcards-extra7` を追加する。KV namespace 6 つと repo variables `WORKER_EXTRA6_URL` / `WORKER_EXTRA6_STAGING_URL` / `WORKER_EXTRA7_URL` / `WORKER_EXTRA7_STAGING_URL` はコミット前に作成済みである（ADR-0036 の制約どおり）。

### casino の溶接を解く

ADR-0038 では casino の clean ユニットを 0 と記録したが、`GameResult` は既に `game_result.go` に移っていた。残っていた溶接は `teamName`（Tressette.go）と、`combinations` / `compareHighCardsSlice` およびその callee `isWheelHand` / `tieBreakValues`（HoldemPlayer.go）だった。これらをタグなしファイル（`trick_helpers.go` / 新規 `poker_hand_helpers.go`）へ純粋移動した。タグなしファイルは全 Worker に入るが、呼ばれない Worker では TinyGo の DCE が落とすためバイト増はない。これで casino から 16 ゲームを移せた。

`movability.py` の残る偽陽性は、構造体フィールド `dealerIdx` やコメント中の型名（`FortyFives`, `Tute`, `Baccarat`）を参照として数えるものだった。実依存は courtpiece → tarneeb（`isValidSuit`）だけだった。

バケット追加と再バケットは ADR-0037 / ADR-0038 と同じく 1 PR で行う。TinyGo 更新も同じ PR に含め、0.42.0 の数字で直接測る。

## 実測結果

CI の TinyGo 0.42.0 で 2 ラウンド実施した。

### ラウンド 1

82 ゲームを移動した（run 35853900184）。既存 8 Worker はすべて余裕が 100 KB を超える状態に戻ったが、新 Worker の `extra6` は 1,019,441 B（余裕 28.4 KB）、`extra7` は 1,003,132 B（余裕 44.4 KB）となり、容量が詰まった。移動先では gzip/ソース比が約 0.26 に上がり、移動元で空いた分は約 0.20 だった。新しい Worker では既存コードとの共通部分が少なく圧縮が効きにくい、という解釈は推測である。

### ラウンド 2

6 ユニット、11 ゲームを既存 Worker へ戻した（run 35855547375）。最終値:

| Worker | gzip (B) | 余裕 (KB) | ゲーム数 |
|--------|---------:|----------:|---------:|
| casino | 938,600 | 107.4 | 51 |
| classic | 935,932 | 110.0 | 40 |
| solo | 936,647 | 109.3 | 47 |
| extra | 939,538 | 106.5 | 33 |
| extra2 | 929,934 | 115.9 | 41 |
| extra3 | 937,796 | 108.2 | 32 |
| extra4 | 930,083 | 115.7 | 36 |
| extra5 | 938,669 | 107.3 | 31 |
| extra6 | 959,197 | 87.3 | 36 |
| extra7 | 958,906 | 87.6 | 36 |
| **合計** | | **1,055.1** | **383** |

`extra6` / `extra7` は目標の 100 KB に約 13 KB届かない。ゲームの移し替えは出す側で約 0.19、入れる側で約 0.2 の gzip/ソース比となり、ほぼ等価交換だった。完全に均しても平均余裕は約 105 KBなので、ここで移動を打ち切った。

## Consequences

### 良い点

- `errors.As` の根本原因が直った TinyGo に上がった。
- 全 Worker で 87 KB 以上の余裕を確保した。
- casino の溶接が解けた。

### 悪い点・受け入れるリスク

- 固定費 318.6 KB × 10 = 約 3.2 MB が重複する。
- Worker が変わった 78 ゲームでは進行中の KV セッションが失われる（ラウンド 1 の 82 ゲームのうち、kingo / oichokabu / tusac / zheng の 4 つはラウンド 2 で元の Worker に戻ったので失われない）。各 Worker の `GAME_SESSIONS` は自身の namespace のみを参照する（`internal/infrastructure/worker/register.go`）。
- `extra6` / `extra7` の余裕は小さいため、新規ゲームは `extra2` / `extra4` から入れる。
- `TestNoErrorsAsAnywhereUnderInternal` は残す。0.42.0 で `errors.As` が安全かは本番で確認していない。外す場合は別 PR とする。

## References

- [ADR-0027](0027-cloudflare-workers-wasm.md)
- [ADR-0028](0028-kv-session-persistence.md)
- [ADR-0032](0032-fourth-worker-capacity.md)
- [ADR-0036](0036-fifth-sixth-worker-capacity.md)
- [ADR-0037](0037-seventh-worker-capacity.md)
- [ADR-0038](0038-eighth-worker-capacity.md)
- [PR #8003](https://github.com/yuta-yoshinaga/go_trumpcards/pull/8003)
- [docs/cloudflare-workers.md](../cloudflare-workers.md)
