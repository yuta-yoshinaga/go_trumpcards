# Contributing to go_trumpcards

go_trumpcards へのコントリビューションに興味を持っていただきありがとうございます！このガイドでは、コントリビューションの手順を説明します。

## Getting Started

### 前提条件

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [Bun](https://bun.sh/) | 1.3.10 |
| [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) | latest |
| [golangci-lint](https://golangci-lint.run/) | latest |

### セットアップ

```sh
git clone https://github.com/yuta-yoshinaga/go_trumpcards.git
cd go_trumpcards
cd frontend && bun install && cd ..
```

## Development Workflow

### 1. Issue を確認する

- 既存の [Issue](https://github.com/yuta-yoshinaga/go_trumpcards/issues) を確認し、作業したいものがあればコメントしてください
- 新しい機能やバグ修正を提案する場合は、まず Issue を作成してください

### 2. ブランチを作成する

```sh
git checkout develop
git pull origin develop
git checkout -b <type>/<short-description>
```

ブランチ名の例: `feat/new-game`, `fix/blackjack-bug`, `docs/update-readme`

### 3. コードを書く（TDD）

このプロジェクトではテスト駆動開発を採用しています:

1. **Red** — 失敗するテストを先に書く
2. **Green** — テストをパスする最小限のコードを書く
3. **Refactor** — テストを維持しながらコードを整理する

### 4. コミット前チェック

すべてのチェックが通ることを確認してください:

**Backend (Go)**

```sh
goimports -w ./...                  # フォーマット
golangci-lint run ./...             # Lint
go test -tags test -p 2 ./...      # テスト
```

**Frontend (React)**

```sh
cd frontend
bun run build   # ビルド
bun run check   # Biome lint + フォーマットチェック
bun run test    # ユニットテスト
```

### 5. コミットする

[Conventional Commits](https://www.conventionalcommits.org/) 形式に従ってください:

```
<type>: <description>
```

| Type | 用途 |
|------|------|
| `feat` | 新機能 |
| `fix` | バグ修正 |
| `docs` | ドキュメントのみの変更 |
| `test` | テストの追加・修正 |
| `refactor` | リファクタリング |
| `perf` | パフォーマンス改善 |
| `chore` | ビルドプロセスや補助ツールの変更 |
| `ci` | CI/CD の変更 |

### 6. Pull Request を作成する

- PR のターゲットブランチは `develop` です
- 関連する Issue がある場合は `Closes #123` で紐付けてください
- CI（lint, test, CodeQL）がすべて通ることを確認してください

## Architecture

Clean Architecture を採用しています。依存の方向は外側から内側への一方向です。

```
infrastructure → adapter → usecase → domain
```

詳細は [docs/architecture.md](docs/architecture.md) を参照してください。

## Quality Standards

### カバレッジ

- **Go**: `cmd/` と `internal/infrastructure/` を除く全パッケージでブランチカバレッジ (C1) 100%
- **Frontend**: `src/{api,components,pages,utils}` でブランチカバレッジ (C1) 100%

### ドキュメント

- 新しいエクスポートされたシンボルには GoDoc / TSDoc コメントを付けてください
- コード変更に伴うドキュメント更新については [CLAUDE.md の Documentation Maintenance](CLAUDE.md#documentation-maintenance) を参照してください

## Adding a New Game

新しいゲームを追加する場合は、[CLAUDE.md の New Game Addition Checklist](CLAUDE.md#new-game-addition-checklist) に従ってください。バックエンド、フロントエンド、ドキュメントの全ステップを1つの PR で完了させてください。

## Infrastructure Notes

### GitHub Pages (Repomix)

`master` ブランチへのマージ時に、GitHub Actions が [repomix](https://github.com/yamadashy/repomix) を実行し、リポジトリスナップショットを GitHub Pages にデプロイします:

```
https://yuta-yoshinaga.github.io/go_trumpcards/repomix-output.txt
```

**初回セットアップ（リポジトリ管理者向け）:** Settings > Pages で Source を **GitHub Actions** に設定してください。

## Questions?

不明な点があれば、Issue でお気軽にご質問ください。
