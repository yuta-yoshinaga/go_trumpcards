# ADR-0027: Cloudflare Workers (TinyGo/Wasm) によるエッジデプロイ

## Status

Accepted

## Date

2026-03-28

## Context

既存のデプロイ先である Render（Docker）は無料枠でコンテナがスリープし、初回アクセス時に30〜60秒のコールドスタートが発生していた。ユーザー体験を損なうこの遅延を解消するため、エッジコンピューティング基盤への移行を検討した。

Cloudflare Workers はリクエスト駆動でコールドスタートが事実上なく、無料枠（1日10万リクエスト）で十分な規模をカバーできる。ただし、無料枠のデプロイサイズ上限は **圧縮後1MB** であり、Go標準コンパイラで生成した Wasm バイナリ（26ゲーム全体で約11MB、gzip後約2.9MB）ではこの制限を超過する。

## Decision

以下の構成を採用した:

### 1. TinyGo による Wasm コンパイル

Go 1.26 の代わりに TinyGo 0.40.1（Go 1.25 対応）を使用し、Wasm バイナリサイズを大幅に削減する。`go.mod` に `go 1.25.8` + `toolchain go1.26.0` を指定することで、Docker ビルド（Go 1.26）と Workers ビルド（TinyGo/Go 1.25）の両方を同一リポジトリで維持する。

### 2. 3-Worker 分割構成

26ゲームをフロントエンドの既存カテゴリに合わせて3つの Worker に分割:

| Worker | カテゴリ | ゲーム数 | gzip後サイズ |
|--------|---------|---------|-------------|
| casino | テーブル + ポーカー | 10 | ~565 KB |
| classic | トリックテイキング + マッチング | 9 | ~589 KB |
| solo | ソリティア + ラミー | 7 | ~524 KB |

全 Worker が無料枠の1MB制限に収まる。

### 3. フロントエンドの Worker URL ルーティング

`gameApi.ts` にゲームごとの Worker URL マッピングを追加。環境変数 `VITE_WORKER_{CASINO,CLASSIC,SOLO}_URL` が未設定の場合は相対URLにフォールバックし、Docker デプロイとの互換性を維持する。

### 4. CI/CD パイプライン

- `develop` マージ → ステージング環境にデプロイ（Worker名に `-staging` サフィックス）
- `master` マージ → 本番環境にデプロイ
- サイズチェック: gzip後1MB超で CI 失敗

### 検討した代替案

| 案 | 不採用理由 |
|----|-----------|
| 標準Go Wasm（有料枠$5/月） | gzip後2.9MBで無料枠に収まらない。有料枠（5MB）なら可能だが、無料枠を優先した |
| TinyGo 単一Worker | 26ゲーム全体だとgzip後1MBを超える可能性が高い |
| Fly.io 等のコンテナPaaS | コールドスタート問題が残る |
| TinyGo対応を待ってGo 1.26維持 | TinyGoのGo 1.26サポートが未定。`go.mod` のデュアルバージョン戦略で即座に対応可能だった |

## Consequences

- **コールドスタート解消**: Workers はリクエスト駆動のため、Docker PaaS のようなスリープ→起動の遅延がない
- **go-json-rest 依存の除去**: TinyGo 互換性のために `go-json-rest` を標準 `net/http` に置き換えた（PR #1012）。結果として外部依存が1つ減り、コードも約2,000行削減された
- **モック ファイルのビルドタグ**: 全モックファイルに `//go:build test` タグを追加。TinyGo ビルドから `testify/mock` を除外するために必要だった
- **Go バージョン制約**: TinyGo が Go 1.25 までしかサポートしないため、`go.mod` に `go 1.25.8` を指定。`toolchain go1.26.0` ディレクティブにより通常開発は Go 1.26 を使用
- **メソッドプレフィックスルーティング不可**: TinyGo の `net/http` は Go 1.22 の `"POST /path"` パターンを未サポート。Workers エントリポイントでは通常の `"/path"` パターンを使用
- **セッション管理**: `SessionProvider[T]` インターフェースにより、Docker 版（インメモリ `MemorySessionProvider`）と Workers 版（`KVSessionProvider` + Cloudflare KV）を統一的に扱う。KV 版はゲームドメインオブジェクトを JSON シリアライズして永続化し、TTL=1時間で自動削除される。KV 無料枠（書き込み 1,000回/日）のためデモ用途に限定。ゲームごとに段階的にシリアライズを追加する方式（Phase 1 は Baccarat のみ対応）
