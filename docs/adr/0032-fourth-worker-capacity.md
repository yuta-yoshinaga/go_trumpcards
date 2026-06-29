# ADR-0032: 4 つ目の Cloudflare Worker（容量バケット）の追加

## Status

Accepted

## Date

2026-06-29

## Context

ゲームは TinyGo WASM バイナリとして 3 つの Cloudflare Worker（`casino` / `classic` / `solo`）に分割デプロイされている。各 Worker は Cloudflare Workers の無料枠 **1 MB gzip** 上限に収める必要があり、`Category` は「ユーザー向けの分類」ではなく純粋に **バイナリサイズのバケット**である（[ADR-0027](0027-cloudflare-workers-wasm.md)）。

179 ゲーム時点で 3 Worker はいずれも上限間際に達した（2026-06-29 計測, gzip）:

| Worker | gzip | 余裕 |
|--------|------|------|
| solo | 1,047,593 | **983 B** |
| casino | 1,046,639 | ~1.9 KB |
| classic | 1,029,181 | ~19 KB |

新規ゲーム候補バッチ（#2880–#2924）の残り約 30 ゲームの大半が `solo` / `classic` / `casino` を必要とするため、**容量上限が新規ゲーム追加のブロッカー**になった。代替案の検討:

1. **コードサイズ削減**（TinyGo フラグ調整・共通化）— 数 KB の改善は見込めるが、約 30 ゲーム分（数百 KB）には届かない見込み。一時しのぎ。
2. **既存 Worker のさらなる分割なしに諦める** — 残りバッチを実装できない。
3. **4 つ目の Worker を追加し、4 バケットへ再バランス** — 各 Worker に十分な余裕（~200 KB+）を確保でき、残りバッチを継続できる。

## Decision

4 つ目の size バケット **`CategoryExtra`（Worker 名 `extra` / `go-trumpcards-extra`）** を追加する。`Category` が size バケットである原則は変わらない。既存 3 Worker から overflow 分のゲームを `extra` へ移し、4 Worker をそれぞれ ~800 KB 程度（~200 KB 以上の余裕）に再バランスする。

ゲームを別 Worker へ移す具体的手順（1 ゲームあたり）:

1. 当該ゲームの全ソースファイルの build tag を `//go:build !js || !wasm || <old>` → `|| extra` に変更（汎用ヘルパーは build tag なしのため全 Worker に含まれ、変更不要）。
2. `RegisterKVGame` 呼び出しを `<old>/<old>.go` から `extra/extra.go` へ移し、`games.Category<Old>` → `games.CategoryExtra` に変更。
3. `registry.go` の `{Name, Category}` を更新。
4. `frontend/src/api/gameApi.ts` の `workerUrl` を `WORKER_EXTRA` に更新。
5. `registry_test.go` のカテゴリ別カウント assertion を更新。

TinyGo のバイナリサイズはローカルで計測できないため、再バランスは CI の `Cloudflare Workers Build` ジョブの size レポートを見ながら段階的に調整する。

## Consequences

- **メリット**: 残り約 30 ゲームを継続実装・デプロイできる。各 Worker に将来分の余裕も確保。
- **デプロイ**: `extra` Worker 用に新しい KV namespace を作成し（`wrangler kv namespace create GAME_SESSIONS`）、`workers/extra/wrangler.toml` の id を設定、`make deploy-workers`（または `wrangler deploy`）でデプロイする。フロントエンドの `VITE_WORKER_EXTRA_URL` 環境変数も設定する。
- **コスト**: ゲーム移動はファイル数が多く機械的（1 ゲーム ~9 ファイルの build tag 変更 + 登録移動）。共通の category 固有ヘルパー（例: `betting.go` は `|| casino`）を使うゲームを移す場合は、そのヘルパーにも `|| extra` を追加する必要がある。
- **段階導入**: 本 ADR の最初のコミットは **インフラのスキャフォールドのみ**（空の `extra` Worker + 4 バケット対応の registry / CI / Makefile / wrangler）。ゲームの移動は CI のサイズ計測を見ながら後続コミットで段階的に行う。
