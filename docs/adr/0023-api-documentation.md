# ADR-0023: GoDoc/TSDoc + GitHub PagesによるAPIドキュメント自動生成

## Status

Accepted

## Date

2026-03-20

## Context

ソースコードのドキュメントと閲覧可能なAPIリファレンスが分離していた。開発者やAIアシスタントがGoパッケージやTypeScriptモジュールの一元的な自動生成ドキュメントを参照する手段がなかった。16以上のゲーム実装を持つバックエンド・フロントエンドにおいて、手動でのドキュメント保守はスケールしない。

必要だったもの:
1. すべてのエクスポートされたシンボルへのインラインドキュメントコメント（GoはGoDoc、TypeScriptはTSDoc）
2. 閲覧可能なHTMLドキュメントの自動生成
3. リリースごとに公開サイトへの自動デプロイ

## Decision

以下のドキュメントスタックを採用する:

- **GoDocコメント**（`// SymbolName description`）を`internal/`配下のすべてのエクスポートされたGoシンボルに付与
- **TSDocコメント**（`/** description */`）を`frontend/src/`配下のすべてのエクスポートされたTypeScriptシンボルに付与
- **gomarkdoc**でGoパッケージからMarkdownを生成し、**pandoc**でHTMLに変換
- **TypeDoc**でTypeScriptソースからHTMLドキュメントを生成
- **GitHub Pages**デプロイをGitHub Actionsで`master`へのpush時に実行（既存のrepomixデプロイと統合）

生成されるサイト構成:
```
_site/
  index.html          # リンク付きランディングページ
  repomix/             # リポジトリスナップショット（NotebookLM用に分割）
  go/                  # Go APIドキュメント（パッケージごとのHTML）
  ts/                  # TypeScript APIドキュメント（HTML）
```

## Consequences

**メリット:**

- すべてのエクスポートされたシンボルにインラインドキュメントが付き、IDEホバー情報とコードの可読性が向上
- 自動生成ドキュメントがコードと同期し続ける（手動でのHTML保守が不要）
- 単一のワークフローが以前のrepomix専用デプロイを置き換え
- 開発者とAIアシスタントがGitHub PagesのURLでAPIドキュメントを閲覧可能

**デメリット:**

- 約400ファイルへのGoDoc/TSDoc追加は大きな初期投資
- CIビルド時間がわずかに増加（gomarkdoc + pandoc + TypeDoc生成）
- TypeDocがdev dependencyとして追加され、`node_modules`サイズがわずかに増加

**その他:**

- ドキュメント品質はコメント品質に依存する（自動生成は正確性を検証しない）
- `deploy-repomix.yml`ワークフローは統合された`deploy-pages.yml`に置き換えられ削除
