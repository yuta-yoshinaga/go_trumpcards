---
trigger: always_on
---
# Frontend Development Rules

Frontend 実装時の主要ルール。詳細は [`CLAUDE.md`](../../CLAUDE.md) および [`frontend/CLAUDE.md`](../../frontend/CLAUDE.md) を参照。

- **パッケージマネージャ**: `npm` / `npx` / `node ./node_modules/...` は使用禁止。`bun` と `bunx` のみ使用する。
- **検証コマンド（コミット前に全通過必須）**:
  - `bun run build`
  - `bun run check` (biome + 各種ガードスクリプト)
  - `bun run typecheck` (TypeScript 7 型検査)
  - `bun run test` (Vitest 単体試験)
- **型検査の必須性**: `bun run typecheck` は必須。素の `tsc` は 5.9 で別物。`vitest` も `biome` も型検査を行わない。
- **テストフレームワーク**: Vitest + React Testing Library。ページ試験とフック試験は `renderWithProviders` ([`frontend/src/test/renderWithProviders.tsx`](../../frontend/src/test/renderWithProviders.tsx)) を使用する。
- **フィクスチャの手組み禁止**: フィクスチャを手で組まない。既存の `makeXxxState()` 等を土台に差分だけ上書きする。手組みは型が変わったときに黙ってずれる（存在しないフィールドで組まれた事故の実例あり）。
- **カバレッジ基準**: `src/{api,components,pages,utils}` は分岐カバレッジ (C1) 80% 以上必須。
