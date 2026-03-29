# Desktop UI/UX Review

PC画面（1280x800, 一般的なラップトップ相当）で全ゲーム画面のスクリーンショットを撮影し、UI/UX課題を抽出してGitHub Issueを作成する。

## 前提条件

- Playwright のChromium がインストール済み（`~/.cache/ms-playwright/`）
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

### 2. スクリーンショット撮影

HashRouter (`/#/path`) を使用。全ゲーム画面のフルページスクリーンショットを撮影する。

**2a. 初期表示状態の撮影（チュートリアルダイアログ付き）**

```sh
mkdir -p /tmp/desktop-screenshots
# ゲーム一覧を gameRoutes.ts から動的に取得（信頼できるソース）
ROOT_GAME=$(grep -P "path:\s*'/'" frontend/src/constants/gameRoutes.ts | grep -oP "labelKey:\s*'nav\.\K[^']*")
ALL_GAMES=$(grep -oP "path:\s*'/\K[^']*" frontend/src/constants/gameRoutes.ts | sed "s/^$/$ROOT_GAME/" | tr '\n' ' ' | sed 's/ $//')
GAMES="${ARGUMENTS:-$ALL_GAMES}"
for game in $GAMES; do
  path="/#/$game"
  [ "$game" = "$ROOT_GAME" ] && path="/"
  PLAYWRIGHT_BROWSERS_PATH=~/.cache/ms-playwright bunx playwright screenshot \
    --browser chromium --viewport-size "1280,800" --full-page \
    --wait-for-selector '[aria-live="polite"]' \
    "http://localhost:8080$path" "/tmp/desktop-screenshots/${game}.png" 2>&1
done
```

**2b. チュートリアルスキップ後のプレイ画面撮影**

Playwrightスクリプトを使い、チュートリアルをスキップしてからプレイ状態のスクリーンショットを撮る。ゲーム操作（ベット、カード選択等）も可能な場合は行う。

```js
// /tmp/desktop-play-screenshots.js
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch({ headless: true });
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const page = await ctx.newPage();

  async function dismissTutorial(page) {
    const skipBtn = page.getByRole('button', { name: 'スキップ' });
    if (await skipBtn.isVisible({ timeout: 1500 }).catch(() => false)) {
      await skipBtn.click();
      await page.waitForTimeout(300);
    }
  }

  // ゲーム一覧を gameRoutes.ts から動的に取得（信頼できるソース）
  const fs = require('fs');
  const src = fs.readFileSync('frontend/src/constants/gameRoutes.ts', 'utf8');
  const rootLabel = src.match(/path:\s*'\/'\s*,\s*labelKey:\s*'nav\.([^']*)'/);
  const rootGame = rootLabel ? rootLabel[1] : 'blackjack';
  const allGames = [...src.matchAll(/path:\s*'\/([^']*)'/g)].map(m => ({
    name: m[1] || rootGame,
    path: m[1] ? `/#/${m[1]}` : '/',
  }));

  // $ARGUMENTS で指定されたゲームのみにフィルタ（未指定時は全ゲーム）
  const args = process.env.REVIEW_GAMES;
  const games = args
    ? allGames.filter(g => args.split(/\s+/).includes(g.name))
    : allGames;

  for (const { name, path } of games) {
    await page.goto(`http://localhost:8080${path}`);
    // Wait for game content to render (skeleton disappears, real content with aria-live appears)
    await page.waitForSelector('[aria-live="polite"]', { timeout: 10000 });
    await dismissTutorial(page);
    await page.waitForTimeout(300);
    await page.screenshot({ path: `/tmp/desktop-screenshots/${name}-play.png`, fullPage: true });
  }

  await browser.close();
})();
```

実行: `PLAYWRIGHT_BROWSERS_PATH=~/.cache/ms-playwright REVIEW_GAMES="$ARGUMENTS" bun /tmp/desktop-play-screenshots.js`

**2c. ナビゲーションメニュー展開状態の撮影**

PC版ではサイドバーまたは展開メニューとして表示される可能性がある。

```js
// /tmp/desktop-nav-screenshot.js
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage({ viewport: { width: 1280, height: 800 } });
  await page.goto('http://localhost:8080/');
  // Wait for game content to render
  await page.waitForSelector('[aria-live="polite"]', { timeout: 10000 });
  const skipBtn = page.getByRole('button', { name: 'スキップ' });
  if (await skipBtn.isVisible({ timeout: 1500 }).catch(() => false)) {
    await skipBtn.click();
    await page.waitForTimeout(300);
  }
  // ハンバーガーメニューが存在する場合はクリック（PC版ではメニューバーかもしれない）
  const menuBtn = page.getByRole('button', { name: 'メニューを開く' });
  if (await menuBtn.isVisible({ timeout: 1000 }).catch(() => false)) {
    await menuBtn.click();
    await page.waitForTimeout(500);
  }
  await page.screenshot({ path: '/tmp/desktop-screenshots/nav-open.png', fullPage: true });
  await browser.close();
})();
```

### 3. スクリーンショット確認

Read ツールで各スクリーンショットを読み込み、視覚的に確認する。

#### 確認観点（最重要: スクロールなしで遊べるか）

最優先の観点は**ビューポート内にゲームのすべての操作・情報が収まり、スクロールなしでストレスなく遊べるか**。デスクトップでも同様に、プレイ中にスクロールが必要な状態は課題である。

**PC固有の観点**

- **ビューポート収まり**: プレイ中の全要素（カード、ボタン、スコア等）が1280x800のビューポート内に収まるか。縦スクロール・横スクロールともに発生しないことが理想
- **ワイドスクリーン活用**: 横幅を活かしてビューポート内に全要素を無理なく配置できているか。ただし「空白を埋めること」自体を目的にしない
- **マウスインタラクション**: hover状態の視覚フィードバック、ドラッグ&ドロップ対応の有無
- **キーボードアクセシビリティ**: Tab移動、Enter/Space操作、ショートカットキーの有無
- **ナビゲーション**: PC版でもハンバーガーメニューのままか。サイドバーやトップメニューバーが適切か

**共通の観点**

- **チュートリアルダイアログ**: 背景との干渉、画面中央配置の適切さ
- **カード表示**: 背景とのコントラスト、視認性。ビューポートに収まる範囲でのサイズ調整
- **テキスト**: フォントサイズの適切さ、行長の読みやすさ。ビューポート内に収めるための情報密度の最適化
- **ボタン/クリック対象**: クリック可能であることの視覚的手がかり。ただしサイズを大きくしてスクロールが発生するなら小さいほうが良い
- **レイアウト**: ビューポート内にすべてが収まるか。収まらない場合、何を削るか・折りたたむかの優先順位
- **デザイン一貫性**: 背景色、ボタンスタイル、テーマ統一

#### 設計原則による判断基準

**「そのデバイスでスクロールせずに快適に遊べるか」**が唯一最大の判断軸。症状ではなく、以下の設計原則で課題かどうかを判断する:

- **スクロールゼロが最優先**: ゲームプレイ中にスクロールが必要な状態は原則として課題。要素を大きくする、情報を追加する等の「改善」がスクロールを生むなら、それは改悪
- **ビューポートに収めるための妥協は許容**: ボタンやカードが小さくても、ビューポート内に全要素が収まりゲームが成立するならOK。逆に大きく見やすくてもスクロールが必要なら問題
- **参照情報はオンデマンド**: 配当表・ルール説明等の参照情報は折りたたみ表示が原則。展開して空白を埋めるのは本末転倒
- **アクション誘導**: ユーザーが次のアクションを迷わずできるか。ただしCTAの明確さのためにスクロールが発生する設計は不可
- **一貫性 > 個別最適**: 同種のゲーム（例: Video Poker系3ゲーム）は同じレイアウトパターンを共有すべき

### 4. 既存Issue確認（起票前に必須）

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

### 5. 課題分類

発見した課題を以下の重大度で分類する:

| 重大度 | 基準 |
|--------|------|
| HIGH | ゲームプレイ不能、またはプレイ中にスクロールが必須 |
| MEDIUM | スクロールなしで遊べるが操作性や視認性に問題がある |
| LOW | 見た目の統一感、細かい改善点（ゲームプレイに支障なし） |

### 6. GitHub Issue作成

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
Desktop (1280x800)
```

### 7. スクリーンショット添付

catbox.moe に画像をアップロードし、各IssueにコメントでMarkdown画像リンクを添付する。

```sh
# アップロード
URL=$(curl -s -F "reqtype=fileupload" -F "fileToUpload=@/tmp/desktop-screenshots/game.png" https://catbox.moe/user/api.php)

# Issueコメントに添付
gh issue comment <ISSUE_NUMBER> --body "## スクリーンショット（Desktop 1280x800）

### ゲーム名
![ゲーム名]($URL)"
```

### 8. クリーンアップ

```sh
pkill -f 'go run' || true
rm -rf /tmp/desktop-screenshots /tmp/desktop-play-screenshots.js /tmp/desktop-nav-screenshot.js
```

## 注意事項

- **リソース制約**: WSL2環境（~2GB RAM）のため、サーバー起動中に他の重いタスク（go test, bun run test等）は実行しない
- **HashRouter**: URLは `http://localhost:8080/#/<game>` 形式。`/` 直打ちは404になる
- **チュートリアルダイアログ**: 初回表示時にほぼ全ゲームで表示される。スキップ状態はlocalStorageに保存される
- **ヘッドレスChromium**: `/opt/google/chrome/chrome` ではなく `~/.cache/ms-playwright/` のChromiumを使う。`PLAYWRIGHT_BROWSERS_PATH` 環境変数で指定する
- **catbox.moe**: 永続的な無料ホスティング。APIキー不要。1ファイル200MBまで
- ゲームルート一覧は `frontend/src/constants/gameRoutes.ts` で管理されており、本コマンドは同ファイルから動的に取得するため、ゲームの追加・削除時にこのコマンドファイルの更新は不要
- モバイル版のレビュー結果と照合し、PC・モバイル共通の課題は既存Issueにデスクトップのスクリーンショットを追記する（重複Issue作成を避ける）

## 引数

$ARGUMENTS — オプション。特定のゲーム名を指定すると、そのゲームのみレビューする（例: `blackjack holdem`）。省略時は全ゲーム。
