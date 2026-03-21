# ADR-0025: Migrate frontend test runner from Vitest to bun:test

## Status

Accepted

## Date

2026-03-21

## Context

フロントエンドテストランナーとしてVitestを使用していたが、ADR-0021でパッケージマネージャーをBunに移行済みであり、Bun 1.3.10が提供する組み込みテストランナー（`bun:test`）はVitest互換の`vi`オブジェクトをサポートしている。Vitest + @vitest/coverage-v8の依存を削除し、Bunファースト戦略に統一することで依存数を削減できる。

## Decision

フロントエンドテストランナーをVitestからbun:testに移行する:

- **import変更**: `from 'vitest'` → `from 'bun:test'`（115ファイル）
- **vi.mocked() 置換**: bun:testには`vi.mocked()`がないため、`asMocked<T>()`ヘルパー（`src/test/viCompat.ts`）で代替
- **vi.mock() → vi.spyOn()**: bunのmodule mock（`vi.mock`/`mock.module`）はプロセス全体でグローバルに適用され、Vitestのようなファイル単位の隔離がない。そのため`vi.spyOn()` + `vi.restoreAllMocks()`パターンに変換し、テスト間の汚染を防止
- **vi.stubGlobal() 置換**: bun:testには`vi.stubGlobal()`がないため、`globalThis`への直接代入 + save/restoreパターンで代替
- **jsdom手動セットアップ**: bun:testはVitestのような`environment: 'jsdom'`設定がないため、`src/test/jsdom-setup.ts`でJSDOMインスタンスを作成しグローバルに登録
- **jest-dom型拡張**: `expect.extend(matchers)`でjest-domマッチャーを登録し、`src/test/jest-dom.d.ts`で型定義を拡張
- **toMatchSnapshot → 明示的アサーション**: bun:testのスナップショットがjsdomノードでハングするため、StatusBadgeテストを明示的なクラス名アサーションに変換
- **依存削除**: `vitest`, `@vitest/coverage-v8`を削除。`bun-types`を追加
- **vite.config.ts**: `vitest/config` import → `vite`に変更、`test`セクション削除
- **bunfig.toml**: テスト設定（preload, timeout）を新規作成
- **テスト実行戦略**: 非ページテストは一括実行、ページテストは個別プロセス実行（module mock汚染とpending promise蓄積を回避）
- **SpadesPageテスト分割**: bun:testがテスト完了後にpending Promiseでプロセスを終了できない問題を回避するため、55テストを3ファイル（27+14+14）に分割

## Consequences

### 利点

- 依存パッケージ2個（vitest, @vitest/coverage-v8）削減
- Bunファースト戦略の統一（ADR-0021の延長）
- テスト起動時間の短縮（Vitestのtransform不要）

### 制約・注意点

- **module mockの非隔離**: `vi.mock()`（`mock.module()`）はプロセス全体に影響するため、`vi.spyOn()` + `restoreAllMocks()`パターンを使用する必要がある。新規テストでも同様
- **pending Promise問題**: bun:testはテスト完了後にpending Promiseがあるとプロセスが終了しない場合がある。React Query + fire-and-forget asyncコールバックの組み合わせで発生しやすい。大量のReactコンポーネントマウント/アンマウントを含むテストファイルは注意が必要
- **ページテスト個別実行**: 上記の理由からページテストは`for f in src/pages/*.test.tsx; do bun test "$f"; done`で個別プロセスとして実行する
- **waitForタイムアウト**: bun上ではsetTimeoutの精度がVitestより低い場合があり、リプレイアニメーション等の実時間待ちテストはタイムアウトを余裕を持って設定する必要がある
- **toFake型未サポート**: `vi.useFakeTimers({ toFake: [...] })`の型定義がbun-typesにないため、`jest-dom.d.ts`で拡張している
