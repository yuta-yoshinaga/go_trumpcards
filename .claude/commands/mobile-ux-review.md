# Mobile UI/UX Review

スマホ画面（375x667, iPhone SE相当）で全ゲーム画面をagent-browserで操作・撮影し、UI/UX課題を抽出してGitHub Issueを作成する。

## 前提条件

- [agent-browser](https://github.com/vercel-labs/agent-browser) がインストール済み（`bun install -g agent-browser && agent-browser install`）
- Go サーバーが起動可能な状態
- `gh` CLIでGitHubにログイン済み
- 画像の置き場所は secret gist（`gh` がログイン済みなら追加設定不要）
- agent-browser の daemon は close するまで Chrome ごと残り続ける。1 回のレビューでは `--session` 名を 1 つに固定し（コマンドごとに名前を変えると daemon がその数だけ増える）、終了時に必ず `agent-browser close` する。保険として `AGENT_BROWSER_IDLE_TIMEOUT_MS=600000` を環境に入れておくと、10 分操作が無い daemon は自動で終了する。

## 手順

### 1. 環境準備

Goサーバーをバックグラウンドで起動する。他セッション・委譲のプロセスには触れず、このコマンドが起動したサーバーだけを停止する。

```sh
[ -f /tmp/mobile-ux-review-server.pid ] && kill "$(cat /tmp/mobile-ux-review-server.pid)" 2>/dev/null || true
PORT=8080 go run ./cmd/server &
echo $! > /tmp/mobile-ux-review-server.pid
# サーバー起動待ち（最大10秒）
for i in $(seq 1 10); do curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/ | grep -q 200 && break; sleep 1; done
```

### 2. ゲーム一覧の取得

```sh
mkdir -p /tmp/mobile-screenshots
# ゲーム一覧を gameRoutes.ts から動的に取得（信頼できるソース）
ROOT_GAME=$(grep -A1 -P "path:\s*'/'," frontend/src/constants/gameRoutes.ts | grep -oP "labelKey:\s*'nav\.\K[^']*")
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

### 4'. 数値計測

スクリーンショット確認の前に、プレイエリアのはみ出しを数値で絞り込む。サーバー起動後に `.claude/commands/scripts/play-area-overflow.sh` を実行する（例: `bash .claude/commands/scripts/play-area-overflow.sh -p 8080 -n 5 -j 4`、ゲーム指定も可能）。配り依存が大きいため、モバイルは複数回リロードした最悪値を使う。目安は mobile ≤300px / desktop ≤600px。`frontend/e2e/play-area-budget.spec.ts` も同じ指標で上限を固定している。`mobile_footer` はフッター内スクロールで隠れる量を示し、目安は ≤40px（手札は見えている必要がある）。 `hidden_cards` は上端 52px（数字の隅と 44px のタップ領域）が `game-footer-content` の表示域に入らない札の数で目安は 0（`frontend/e2e/hand-visibility.spec.ts` と同じ判定）、`footer_cards` は検出できた札数で、0 の場合は hidden_cards の検査結果に意味がない。計測前に、ポートを listen しているサーバーの cwd がこのチェックアウトと異なれば停止する（別ツリーの `public/` を測ってしまうため。意図して測るときは `-x`）。`public/index.html` より新しいファイルが `frontend/src` にあれば、ビルドが古いという警告を出す。

### 4. スクリーンショット確認

Read ツールで各スクリーンショットを読み込み、視覚的に確認する。

#### 確認観点（最重要: スクロールなしで遊べるか）

最優先の観点は**ビューポート内にゲームのすべての操作・情報が収まり、スクロールなしでストレスなく遊べるか**。タップターゲットのサイズを大きくした結果スクロールが発生するなら、それは改善ではなく改悪である。

- **ビューポート収まり**: プレイ中の全要素（カード、ボタン、スコア等）が375x667のビューポート内に収まるか
- **チュートリアルダイアログ**: 背景との干渉、視認性
- **カード表示**: 背景とのコントラスト、視認性。ビューポートに収まる範囲でのサイズ調整
- **テキスト**: 溢れ、切れ。ビューポート内に収めるための情報密度の最適化
- **ボタン/タップ対象**: タップ可能であること。
- **レイアウト**: ビューポート内にすべてが収まるか。収まらない場合、何を削るか・折りたたむかの優先順位
- **デザイン一貫性**: 背景色、ボタンスタイル、テーマ統一
- **ナビゲーション**: ゲーム数に対する探しやすさ

#### 設計原則による判断基準

**「そのデバイスでスクロールせずに快適に遊べるか」**が唯一最大の判断軸。症状ではなく、以下の設計原則で課題かどうかを判断する:

- **収まることが大きさより優先**: タップターゲットを大きくする・情報を足す等の変更がスクロールを生むなら改悪。小さくてもビューポートに収まりゲームが成立するならOK
- **参照情報はオンデマンド**: 配当表・ルール説明等の参照情報は折りたたみ表示が原則。展開して空白を埋めるのは本末転倒
- **アクション誘導**: ユーザーが次のアクションを迷わずできるか
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
| HIGH | ゲームプレイ不能、または操作・情報がビューポートに収まらない |
| MEDIUM | 操作性や視認性に問題がある |
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

影響箇所に挙げたゲームの画像だけを渡す。

```sh
# 最初のIssueでsecret gistを作成（成功後、gist: <id> がstderrに表示される）
bash .claude/commands/scripts/attach-screenshots.sh -i <ISSUE_NUMBER> -t "スクリーンショット（iPhone SE 375x667）" /tmp/mobile-screenshots/bhabhi-play.png /tmp/mobile-screenshots/pitch-play.png

# 2件目以降は同じgistを再利用
bash .claude/commands/scripts/attach-screenshots.sh -i <ISSUE_NUMBER> -t "スクリーンショット（iPhone SE 375x667）" -g <GIST_ID> /tmp/mobile-screenshots/bhabhi-play.png /tmp/mobile-screenshots/pitch-play.png
```

### 9. クリーンアップ

```sh
agent-browser close
[ -f /tmp/mobile-ux-review-server.pid ] && kill "$(cat /tmp/mobile-ux-review-server.pid)" 2>/dev/null || true
rm -f /tmp/mobile-ux-review-server.pid
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

- **HashRouter**: URLは `http://localhost:8080/#/<game>` 形式。`/` 直打ちは404になる
- **チュートリアルダイアログ**: 初回表示時にほぼ全ゲームで表示される。スキップ状態はlocalStorageに保存される
- catbox は使わない: GitHub の画像プロキシ (camo) が 503 を返し、直接 URL は 200 なので投稿前の確認では気づけない (2026-10-14)。スクリプトは投稿後のコメントの `<img src>`（GitHub が実際に読み込む URL）を検証する
- ゲームルート一覧は `frontend/src/constants/gameRoutes.ts` で管理されており、本コマンドは同ファイルから動的に取得するため、ゲームの追加・削除時にこのコマンドファイルの更新は不要

## 引数

$ARGUMENTS — オプション。特定のゲーム名を指定すると、そのゲームのみレビューする（例: `blackjack holdem`）。省略時は全ゲーム。
