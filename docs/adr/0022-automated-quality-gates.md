# ADR-0022: Automated quality gates via Claude Code hooks

## Status

Accepted

## Date

2026-03-20

## Context

git履歴（1253コミット）を分析した結果、繰り返し発生する修正コミットのパターンが特定された:

- **goimportsフォーマット漏れ**: 12回（新規Goファイル作成後に実行忘れ → CI失敗 → 修正コミット）
- **golangci-lint / staticcheck警告**: 12回（lint未実行でコミット → CI失敗 → 修正コミット）
- **biomeフォーマット / import ordering**: 7回（フロントエンドファイル編集後にcheck忘れ）
- **PRレビュー指摘への修正**: 61回（大きなfeatを一度にPR → 多数の指摘 → 修正コミット）
- **コード重複 → 後追いリファクタ**: 71回（新ゲーム追加時にコピー&ペースト → 後から共通化）
- **E2Eテスト不安定**: 12回（ランダム性依存のアサーション、不適切なタイムアウト）
- **デッドコード残留**: 9回
- **カバレッジ不足**: 9回

これらは手動プロセスへの依存が原因であり、自動化で防止可能なものが大半だった。

## Decision

Claude Code hooks（`.claude/settings.json`）を使い、以下の自動品質ゲートを導入する:

### PostToolUse hooks（Write|Edit時に自動実行）

1. **goimports自動実行**: `.go`ファイルをWrite/Editした直後に`goimports -w`を自動実行
2. **biome自動修正**: `.ts`/`.tsx`ファイルをWrite/Editした直後に`bunx biome check --write`を自動実行

### PreToolUse hooks（git commit前にブロッキングチェック）

3. **golangci-lint**: ステージされたGoファイルがある場合、`golangci-lint run ./...`を実行。警告があればコミットをブロック
4. **biome check**: ステージされたTS/TSXファイルがある場合、`bun run check`を実行。エラーがあればコミットをブロック
5. **ドキュメント乖離検知**: agentフックでステージされたファイルからドキュメント更新漏れを検出。CLAUDE.mdのDocumentation Maintenanceテーブルのルールを適用

### ドキュメント・ガイドラインの強化

6. **新ゲーム追加チェックリスト**: CLAUDE.mdに24項目のチェックリストを追加。共通ヘルパーの一覧と、バックエンド/フロントエンド/ドキュメントの全ステップを網羅
7. **E2Eテストガイドライン**: frontend/CLAUDE.mdにflaky testを防ぐための具体的ルールを追記
8. **`/doc-drift-check`スキル**: 既存の乖離を一括で見つけるためのスキル（`.claude/skills/doc-drift-check/`）

## Consequences

**メリット:**

- フォーマット修正コミットの完全排除（goimports 12回、biome 7回 → 0回）
- lint警告の事前検出（golangci-lint 12回 → 0回）
- ドキュメント更新漏れの自動検出
- 新ゲーム追加時の品質底上げ（チェックリストにより重複コードとPRレビュー指摘を削減）
- E2Eテストの安定性向上

**デメリット:**

- PostToolUseフックにより毎回のファイル保存で0.5-1秒のオーバーヘッド
- PreToolUseのgolangci-lintフックでコミット前に最大2分のブロッキング
- agentフック（ドキュメント乖離検知）はLLM呼び出しのためコストが発生
- hookコマンドはツール（`goimports`, `golangci-lint`, `bun`）がPATHに設定されている前提

**リスク軽減:**

- フックは`|| true`で非致命的エラーを無視し、通常の開発フローを阻害しない
- ブロッキングチェック（PreToolUse）はgit commitコマンドのみに限定し、他のBashコマンドはスルー
- `/hooks`メニューからいつでも無効化可能
