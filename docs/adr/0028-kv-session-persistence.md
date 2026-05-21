# ADR-0028: Cloudflare KV によるセッション永続化

## Status

Accepted

## Date

2026-03-28

## Context

ADR-0027 で Cloudflare Workers へのエッジデプロイを採用したが、Workers はステートレスのためリクエスト間でゲームセッション（進行状態・設定・チップ等）が保持されない。Docker 版ではインメモリの `SessionStore[T]` がセッションを管理しているが、Workers 版では毎回新規インタラクターが生成され、ゲーム体験が大きく損なわれていた。

## Decision

以下の構成でセッション永続化を実装した:

### 1. SessionProvider インターフェース

`SessionStore[T]` を直接参照していた `execWithSession` を `SessionProvider[T]` インターフェースに抽象化:

- **MemorySessionProvider**: 既存の `SessionStore` をラップ（Docker 版、変更なし）
- **KVSessionProvider**: Cloudflare KV でセッション状態を読み書き（Workers 版）

Docker 版のコードパスは一切影響を受けない。

### 2. ドメインオブジェクトの JSON シリアライズ

全26ゲームのドメイン型に `MarshalJSON`/`UnmarshalJSON` を実装。JSON フィールド名は短縮形（`"d"`, `"v"`, `"tc"` 等）を使い KV ストレージサイズを最小化。不正入力への防御として全スライスフィールドに `maxSliceLen=1000` のバリデーションを追加。

### 3. KV 環境分離

| 環境 | KV namespace | 用途 |
|------|-------------|------|
| prod | `GAME_SESSIONS` | 本番セッション |
| staging | `GAME_SESSIONS_STAGING` | ステージング |
| local | `GAME_SESSIONS_preview` | ローカル開発 |

wrangler.toml の `[env.staging]` セクションで staging/prod の KV を分離。

### 検討した代替案

| 案 | 不採用理由 |
|----|-----------|
| コマンドリプレイ（イベントソーシング） | シャッフルの非決定性により同じゲーム状態を再現できない（#1654 で再検討する余地はあるが、snapshot 配列方式の方が実装コストが低かった） |
| ドメインフィールドの export | 全26ゲームの構造体フィールドを公開する必要があり、カプセル化を破壊する |
| Durable Objects | 無料枠の制約が KV より厳しく、デモ用途には過剰 |

## Consequences

- **セッション永続化**: Workers 版でも Docker 版と同等のゲーム体験が実現
- **KV 無料枠制約**: 書き込み 1,000回/日のためデモ用途に限定（1アクション=1書き込み）
- ~~**Undo 非永続**: ソリティア系ゲームの Undo 履歴（snapshot）は KV に保存しない。復元後は Undo 不可~~ — #1654 / #1860 で解消済み。Klondike を皮切りに、FreeCell・Spider・TriPeaks・Pyramid・Forty Thieves・Canfield・Yukon・Russian Solitaire・Scorpion・Accordion・Baker's Dozen・Calculation・Golf・Gaps・EightOff・Cruel・Crescent・Seahaven Towers・Spiderette・Beleaguered Castle・Monte Carlo・Poker Squares の全 23 ソリティア／グリッド系ゲームで snapshot を `*SnapshotJSON` の Marshal/Unmarshal で永続化
- **Undo 永続（snapshot 配列方式）**: 当初検討時にイベントソーシングを「シャッフルの非決定性」を理由に却下したが、Klondike の典型 1 ハンドで snapshot 配列が ~150KB（KV 値上限 25 MB に対し十分小さい）に収まる実測を踏まえ、**snapshot 配列をそのまま KV に保存する** 方式を #1654 で採用。各 snapshot は元の `klondikeSnapshot` と同形（タブロー / ストック / ウェイスト / ファンデーション / 進行フラグ）で、`json.Marshaler` 実装を追加して unexported フィールドを露出
- **CPU 記憶永続**: CPU プレイヤーの `memoryManager` (Old Maid / Go Fish / Memory / Doubt) は KV に永続化され、復元後も Hard 難易度の戦略が維持される (#1655 で対応)。データ量が小さいため KV 1MB 制限への影響は無視できる
- ~~**CPU 記憶非永続**: CPU プレイヤーの `memoryManager` は復元時にリセット。CPU は再度学習する~~ — #1655 で解消済み
- **コード量増加**: 全26ゲームに MarshalJSON/UnmarshalJSON を追加（約5,000行）。パターンは統一されており保守性は維持
