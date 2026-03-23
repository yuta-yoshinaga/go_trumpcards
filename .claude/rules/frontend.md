---
globs: ["frontend/**/*.ts", "frontend/**/*.tsx"]
---

# フロントエンド (TypeScript/TSX) ファイル編集ルール

## パッケージマネージャー

**`npm` ではなく `bun` を、`npx` ではなく `bunx` を常に使用すること。** このプロジェクトではBunを唯一のJavaScriptパッケージマネージャー・スクリプトランナーとして使用する。

## コミット前チェック（必須、全て通過すること）

```sh
cd frontend && bun run build   # Reactビルド
cd frontend && bun run check   # Biome lint + フォーマットチェック
cd frontend && bun run test    # Vitestユニットテスト
```

## テスト

**ユニットテストは必須**。実装と同じコミットに含める。テストスタック: **Vitest + React Testing Library + jest-dom**

### TDDサイクル (Red → Green → Refactor)

実装前に必ずこのサイクルを守ること:

1. **Red** — 失敗するテストを先に書く。実装コードを書く前に、期待動作を捉えるテストを作成し失敗を確認する:
   ```sh
   cd frontend && bun run test -- --run TestNewFeature  # 失敗 (Red)
   ```
2. **Green** — テストをパスする最小限のコードを書く。余分な機能は追加しない:
   ```sh
   cd frontend && bun run test -- --run TestNewFeature  # パス (Green)
   ```
3. **Refactor** — テストを維持しながらコードを整理する。命名・構造・重複除去を行い、全テストが通ることを確認:
   ```sh
   cd frontend && bun run test  # 全テストパス (Refactor後)
   ```

### カバレッジ基準

以下4ディレクトリで **ブランチカバレッジ (C1) 80%以上** が必須:

- `frontend/src/api`
- `frontend/src/components`
- `frontend/src/pages`
- `frontend/src/utils`

ビジネスロジックとクリティカルパスに集中してテストすること。到達不能分岐の無理なカバレッジは不要。

### テスト配置（レイヤー別）

| レイヤー | テストファイル | テスト対象 |
|---------|--------------|---------|
| API client | `src/api/*.test.ts` | URL・リクエストボディ・エラーハンドリング |
| Components | `src/components/*.test.tsx` | レンダリング・props・イベントハンドラ |
| Pages | `src/pages/*.test.tsx` | マウント時APIコール・フェーズ別レンダリング・ボタン操作 |

### テストパターン

- **APIモック**: `vi.mock('../api/gameApi', ...)` でAPIモジュールをモック; `vi.mocked(api.exec)` でアクセス
- **ルーター依存コンポーネント**: `NavBar` など `useLocation` を使うコンポーネントは `<MemoryRouter initialEntries={['/path']}>` でラップ
- **非同期エフェクト待機**: `useEffect` でAPIを呼ぶコンポーネントは `waitFor(() => expect(...))` で待つ
- **ボタンのクエリ**: テキストが複数要素に存在する場合は `screen.getByRole('button', { name: '...' })` を使う
- **QueryClientProviderラップ**: pageテスト・hookテストは `renderWithProviders`（`frontend/src/test/renderWithProviders.tsx`）でラップ

## i18n（国際化）

Web GUIはJapanese (ja) / English (en) を `react-i18next` + `i18next-browser-languagedetector` でサポート。

- **設定**: `frontend/src/i18n/index.ts`
- **翻訳ファイル**: `frontend/src/i18n/locales/{ja,en}/<game>.json`
- **コンポーネント内**: `useTranslation()` フックを使う
- **非コンポーネントファイル** (`playerUtils.ts` など): `i18n` インスタンスを直接インポート
- **テスト**: `frontend/src/test/setup.ts` でja翻訳が初期化済み

## デッドコード

- コード変更時に遭遇したデッドコードは必ず削除する
- 検出ツール: `knip`
- 削除前に手動で確認する（静的解析の誤検知に注意）
