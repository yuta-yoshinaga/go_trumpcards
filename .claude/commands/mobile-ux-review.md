# Mobile UI/UX Review

スマホ画面（375x667, iPhone SE相当）で全ゲーム画面をagent-browserで操作・撮影し、UI/UX課題を抽出してGitHub Issueを作成する。

## 前提条件

- [agent-browser](https://github.com/vercel-labs/agent-browser) がインストール済み（`npm install -g agent-browser && agent-browser install`）
- Go サーバーが起動可能な状態
- `gh` CLIでGitHubにログイン済み
- 画像アップロード先として catbox.moe を使用（永続的な無料ホスティング、APIキー不要）

## 手順

### 1. 環境準備

残留プロセスを停止し、Goサーバーをバックグラウンドで起動する。

```sh
pkill -f 'go run' || true; pkill -f vitest || true; pkill -f 'bun run' || true
sleep 1
PORT=8080 go run ./cmd/server &
# サーバー起動待ち（最大10秒）
for i in $(seq 1 10); do curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/ | grep -q 200 && break; sleep 1; done
```

### 2. ゲーム一覧の取得

```sh
mkdir -p /tmp/mobile-screenshots
# ゲーム一覧を gameRoutes.ts から動的に取得（信頼できるソース）
ROOT_GAME=$(grep -P "path:\s*'/'" frontend/src/constants/gameRoutes.ts | grep -oP "labelKey:\s*'nav\.\K[^']*")
ALL_GAMES=$(grep -oP "path:\s*'/\K[^']*" frontend/src/constants/gameRoutes.ts | sed "s/^$/$ROOT_GAME/" | tr '\n' ' ' | sed 's/ $//')
GAMES="${ARGUMENTS:-$ALL_GAMES}"
echo "対象ゲーム: $GAMES"
```

### 3. agent-browserによるスクリーンショット撮影

agent-browser CLIを使い、モバイルビューポートでゲーム画面を撮影・確認する。

**3a. ブラウザ起動・ビューポート設定**

```sh
agent-browser open "http://localhost:8080/"
agent-browser set viewport 375 667
```

**3b. 各ゲームについて以下を繰り返す**

各ゲームについて、以下の操作をagent-browserで実行する:

```sh
for game in $GAMES; do
  path="/#/$game"
  [ "$game" = "$ROOT_GAME" ] && path="/"

  # 1. ページ遷移
  agent-browser open "http://localhost:8080$path"

  # 2. ゲームコンテンツの読み込み待ち
  agent-browser wait '[aria-live="polite"]'

  # 3. 初期表示スクショ（チュートリアルダイアログ付き）
  agent-browser screenshot "/tmp/mobile-screenshots/${game}.png" --full

  # 4. チュートリアルスキップ（スキップボタンがあればクリック）
  agent-browser snapshot
  agent-browser find role button click --name "スキップ" 2>/dev/null || true

  # 5. プレイ画面スクショ
  agent-browser wait 300
  agent-browser screenshot "/tmp/mobile-screenshots/${game}-play.png" --full
done
```

**3c. ナビゲーションメニュー撮影**

```sh
agent-browser open "http://localhost:8080/"
agent-browser wait '[aria-live="polite"]'
agent-browser find role button click --name "スキップ" 2>/dev/null || true
agent-browser wait 300
agent-browser find role button click --name "メニューを開く" 2>/dev/null || true
agent-browser wait 500
agent-browser screenshot "/tmp/mobile-screenshots/nav-open.png" --full
```

### 4. スクリーンショット確認

Read ツールで各スクリーンショットを読み込み、視覚的に確認する。

#### 確認観点（最重要: スクロールなしで遊べるか）

最優先の観点は**ビューポート内にゲームのすべての操作・情報が収まり、スクロールなしでストレスなく遊べるか**。タップターゲットのサイズを大きくした結果スクロールが発生するなら、それは改善ではなく改悪である。

- **ビューポート収まり**: プレイ中の全要素（カード、ボタン、スコア等）が375x667のビューポート内に収まるか。縦スクロール・横スクロールともに発生しないことが理想
- **チュートリアルダイアログ**: 背景との干渉、視認性
- **カード表示**: 背景とのコントラスト、視認性。ビューポートに収まる範囲でのサイズ調整
- **テキスト**: 溢れ、切れ。ビューポート内に収めるための情報密度の最適化
- **ボタン/タップ対象**: タップ可能であること。ただしサイズを大きくしてスクロールが発生するなら小さいほうが良い
- **レイアウト**: ビューポート内にすべてが収まるか。収まらない場合、何を削るか・折りたたむかの優先順位
- **デザイン一貫性**: 背景色、ボタンスタイル、テーマ統一
- **ナビゲーション**: ゲーム数に対する探しやすさ

#### 設計原則による判断基準

**「そのデバイスでスクロールせずに快適に遊べるか」**が唯一最大の判断軸。症状ではなく、以下の設計原則で課題かどうかを判断する:

- **スクロールゼロが最優先**: ゲームプレイ中にスクロールが必要な状態は原則として課題。タップターゲットを大きくする、情報を追加する等の「改善」がスクロールを生むなら、それは改悪
- **ビューポートに収めるための妥協は許容**: カードやボタンが小さくても、ビューポート内に全要素が収まりゲームが成立するならOK。逆に大きく見やすくてもスクロールが必要なら問題
- **参照情報はオンデマンド**: 配当表・ルール説明等の参照情報は折りたたみ表示が原則。展開して空白を埋めるのは本末転倒
- **アクション誘導**: ユーザーが次のアクションを迷わずできるか。ただしCTAの明確さのためにスクロールが発生する設計は不可
- **一貫性 > 個別最適**: 同種のゲーム（例: Video Poker系3ゲーム）は同じレイアウトパターンを共有すべき

### 5. 既存Issue確認（起票前に必須）

課題を発見したら、**起票前に必ず以下を実行**する:

```sh
# 類似の既存issueを検索（open + closed両方）
gh issue list --search "<課題のキーワード>" --state all --limit 20
```

確認事項:
- **同一の課題が過去に起票されていないか**（open/closed問わず）
- **過去に逆方向の対応がされていないか**（例: 「空白を埋める」→「情報過多を減らす」の振り子パターン）
- 矛盾する過去対応がある場合は**新規issueを起票せず**、既存issueにコメントで経緯と根本原因の考察を残す
- 同一課題がclosedで再発している場合は、既存issueを再openするか、根本原因を特定した新issueを起票する

### 6. 課題分類

発見した課題を以下の重大度で分類する:

| 重大度 | 基準 |
|--------|------|
| HIGH | ゲームプレイ不能、またはプレイ中にスクロールが必須 |
| MEDIUM | スクロールなしで遊べるが操作性や視認性に問題がある |
| LOW | 見た目の統一感、細かい改善点（ゲームプレイに支障なし） |

### 7. GitHub Issue作成

各課題ごとにGitHub Issueを作成する。

**ラベル**: `ui/ux` + `bug`（既存の問題）or `enhancement`（改善提案）

**Issue本文テンプレート**:
```markdown
## 問題

[具体的な問題の説明]

### 影響箇所
- [ゲーム名1]: [具体的な症状]
- [ゲーム名2]: [具体的な症状]

## 改善案

1. [改善案1]
2. [改善案2]

## 重大度
[HIGH/MEDIUM/LOW] — [理由]

## 対象デバイス
Mobile (375x667)
```

### 8. スクリーンショット添付

catbox.moe に画像をアップロードし、各IssueにコメントでMarkdown画像リンクを添付する。

```sh
# アップロード
URL=$(curl -s -F "reqtype=fileupload" -F "fileToUpload=@/tmp/mobile-screenshots/game.png" https://catbox.moe/user/api.php)

# Issueコメントに添付
gh issue comment <ISSUE_NUMBER> --body "## スクリーンショット（iPhone SE 375x667）

### ゲーム名
![ゲーム名]($URL)"
```

### 9. クリーンアップ

```sh
agent-browser close
pkill -f 'go run' || true
rm -rf /tmp/mobile-screenshots
```

## agent-browser コマンドリファレンス

本コマンドで使用する主要なagent-browserコマンド:

| コマンド | 用途 |
|----------|------|
| `agent-browser open <url>` | ページ遷移 |
| `agent-browser set viewport 375 667` | ビューポートをモバイルサイズに設定 |
| `agent-browser wait <selector>` | セレクタ出現の待機 |
| `agent-browser wait <ms>` | ミリ秒待機 |
| `agent-browser screenshot <path> --full` | フルページスクリーンショット撮影 |
| `agent-browser snapshot` | アクセシビリティツリー取得（AI向け要素探索） |
| `agent-browser find role button click --name "..."` | ボタンをアクセシビリティ名で検索・クリック |
| `agent-browser find text "..." click` | テキストで要素を検索・クリック |
| `agent-browser click <ref>` | snapshot参照ID（`@e2`等）で要素クリック |
| `agent-browser close` | ブラウザ終了 |

**操作のコツ**:
- `snapshot` でアクセシビリティツリーを取得し、参照ID（`@e1`, `@e2`...）で要素を特定してから `click @e2` で操作する
- `find role button click --name "..."` でアクセシビリティ名ベースの要素探索・操作が可能
- 要素が見つからない場合は `2>/dev/null || true` で握りつぶしてスクリプトを継続する

## 注意事項

- **リソース制約**: WSL2環境（~2GB RAM）のため、サーバー起動中に他の重いタスク（go test, bun run test等）は実行しない
- **HashRouter**: URLは `http://localhost:8080/#/<game>` 形式。`/` 直打ちは404になる
- **チュートリアルダイアログ**: 初回表示時にほぼ全ゲームで表示される。スキップ状態はlocalStorageに保存される
- **catbox.moe**: 永続的な無料ホスティング。APIキー不要。1ファイル200MBまで
- ゲームルート一覧は `frontend/src/constants/gameRoutes.ts` で管理されており、本コマンドは同ファイルから動的に取得するため、ゲームの追加・削除時にこのコマンドファイルの更新は不要

## 引数

$ARGUMENTS — オプション。特定のゲーム名を指定すると、そのゲームのみレビューする（例: `blackjack holdem`）。省略時は全ゲーム。
