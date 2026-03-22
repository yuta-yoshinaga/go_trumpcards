# go_trumpcards
トランプカードゲームアルゴリズムをGoで実装

[![Backend Coverage](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards/graph/badge.svg?flag=backend)](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards)
[![Frontend Coverage](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards/graph/badge.svg?flag=frontend)](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards)

## Vision

**世界中のあらゆるトランプゲームを、誰でも無料で遊べるようにする。**

このプロジェクトは、人間とAIコーディングエージェントが共にソフトウェアを創り上げる **「共創のリファレンスモデル」** です。AIエージェントが正確にコンテキストを理解し、高品質なコードを生成できる開発環境を整備することで、人間とAIの協調開発のベストプラクティスを示し続けます。

go_trumpcardsが目指す未来は、**あらゆる人がクリエイターとなり、自分が欲しいものを生成AIコーディングエージェントとの共創で実現できる世の中** です。

## Description
トランプカードのアルゴリズムをGo+Clean Architectureで実装したプロジェクトです。Web GUIは日本語/英語の多言語対応（react-i18next）をサポートしています。

以下のゲームを実装しています：

- **ブラックジャック (BlackJack)**: CLI および Web GUIで遊べます（チップ/ベッティングシステム、スプリット、ダブルダウン、インシュランス、ソフト17トグル、カードカウンティング練習、マルチプレイヤーCPU席付き） — [CUI版マニュアル](docs/manual/cui/blackjack.md) / [Web版マニュアル](docs/manual/web/blackjack.md)
- **ポーカー (5-card Draw Poker)**: CLI および Web GUIで遊べます（1人 vs CPU×1〜3、4種のプレイスタイル、ジョーカーワイルド、サイドポット対応、キッカー表示） — [CUI版マニュアル](docs/manual/cui/poker.md) / [Web版マニュアル](docs/manual/web/poker.md)
- **ババ抜き (Old Maid)**: CLI および Web GUIで遊べます — [CUI版マニュアル](docs/manual/cui/oldmaid.md) / [Web版マニュアル](docs/manual/web/oldmaid.md)
- **大富豪 (Daifugo)**: CLI および Web GUIで遊べます — [CUI版マニュアル](docs/manual/cui/daifugo.md) / [Web版マニュアル](docs/manual/web/daifugo.md)
- **7並べ (Sevens)**: CLI および Web GUIで遊べます（オプションルール: トンネル、ジョーカー、CPU戦略、片側ストップ、ジョーカー連続禁止） — [CUI版マニュアル](docs/manual/cui/sevens.md) / [Web版マニュアル](docs/manual/web/sevens.md)
- **ダウト (Doubt)**: CLI および Web GUIで遊べます（1人 vs CPU×3、10秒ダウト判定ウィンドウ付き） — [CUI版マニュアル](docs/manual/cui/doubt.md) / [Web版マニュアル](docs/manual/web/doubt.md)
- **テキサスホールデム (Texas Hold'em)**: CLI および Web GUIで遊べます（1人 vs CPU×3、5種のプレイスタイル、サイドポット対応、キッカー表示） — [CUI版マニュアル](docs/manual/cui/holdem.md) / [Web版マニュアル](docs/manual/web/holdem.md)
- **オマハホールデム (Omaha Hold'em)**: CLI および Web GUIで遊べます（1人 vs CPU×3、4枚のホールカード、必ず2枚+3枚でハンドを構成） — [CUI版マニュアル](docs/manual/cui/omaha.md) / [Web版マニュアル](docs/manual/web/omaha.md)
- **ハーツ (Hearts)**: CLI および Web GUIで遊べます（1人 vs CPU×3、トリックテイキング、シュート・ザ・ムーン、3段階CPU難易度） — [CUI版マニュアル](docs/manual/cui/hearts.md) / [Web版マニュアル](docs/manual/web/hearts.md)
- **神経衰弱 (Memory)**: CLI および Web GUIで遊べます（1人 vs CPU×3、52枚のカードから同ランクのペアを揃える、3段階CPU難易度） — [CUI版マニュアル](docs/manual/cui/memory.md) / [Web版マニュアル](docs/manual/web/memory.md)
- **クロンダイク (Klondike)**: CLI および Web GUIで遊べます（1人用ソリティア、52枚のカード、7列のタブロー、山札・ウェスト・4つの組札） — [CUI版マニュアル](docs/manual/cui/klondike.md) / [Web版マニュアル](docs/manual/web/klondike.md)
- **フリーセル (FreeCell)**: CLI および Web GUIで遊べます（1人用ソリティア、52枚のカード、8列のタブロー、4つのフリーセル・4つの組札、スーパームーブ対応） — [CUI版マニュアル](docs/manual/cui/freecell.md) / [Web版マニュアル](docs/manual/web/freecell.md)
- **バカラ (Baccarat)**: CLI および Web GUIで遊べます（プレイヤー/バンカー/タイベット、第3カードルール、5%バンカーコミッション、チップシステム） — [CUI版マニュアル](docs/manual/cui/baccarat.md) / [Web版マニュアル](docs/manual/web/baccarat.md)
- **スペード (Spades)**: CLI および Web GUIで遊べます（1人 vs CPU×3、トリックテイキング＋ビッド、スペードがトランプ、ニルビッド、バッグペナルティ、3段階CPU難易度） — [CUI版マニュアル](docs/manual/cui/spades.md) / [Web版マニュアル](docs/manual/web/spades.md)
- **クレイジーエイト (Crazy Eights)**: CLI および Web GUIで遊べます（1人 vs CPU×3、スートまたはランク一致でカードを出す、8はワイルド、ドローパイル補充、ポイント制マッチ） — [CUI版マニュアル](docs/manual/cui/crazyeights.md) / [Web版マニュアル](docs/manual/web/crazyeights.md)
- **ジンラミー (Gin Rummy)**: CLI および Web GUIで遊べます（1人 vs CPU×1、2人対戦ラミー、10枚の手札、セットとランでメルド、ノック・ジン・アンダーカット、ポイント制マッチ） — [CUI版マニュアル](docs/manual/cui/ginrummy.md) / [Web版マニュアル](docs/manual/web/ginrummy.md)
- **スパイダーソリティア (Spider Solitaire)**: CLI および Web GUIで遊べます（1人用ソリティア、2デッキ104枚、10列のタブロー、難易度3段階（1/2/4スート）、同スート降順シーケンス、完成スート自動除去） — [CUI版マニュアル](docs/manual/cui/spider.md) / [Web版マニュアル](docs/manual/web/spider.md)
- **ナポレオン (Napoleon)**: CLI および Web GUIで遊べます（4人対戦トリックテイキング、52枚+ジョーカー1枚=53枚、絵札ビッド制、ナポレオン軍 vs 連合軍の隠しチーム戦、切り札宣言、副官指名、特殊カード（マイティ・ジョーカーキラー）、3段階CPU難易度） — [CUI版マニュアル](docs/manual/cui/napoleon.md) / [Web版マニュアル](docs/manual/web/napoleon.md)

## Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [Bun](https://bun.sh/) | 1.3.10 |

## Usage
### Install

#### go install（Go ユーザー向け）
```sh
go install github.com/yuta-yoshinaga/go_trumpcards/cmd/trumpcards@latest
trumpcards blackjack
```

#### GitHub Releases からバイナリをダウンロード
Linux/macOS/Windows 向けのビルド済みバイナリは [GitHub Releases](https://github.com/yuta-yoshinaga/go_trumpcards/releases) から入手できます。

**Linux / macOS**
```sh
# VERSION には取得したいバージョン番号を指定（例: v3.12.0）
VERSION=v3.12.0

# OS と アーキテクチャを確認して対応する URL を選択
# Linux amd64:
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_linux_amd64.tar.gz" | tar xz
# Linux arm64:
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_linux_arm64.tar.gz" | tar xz
# macOS amd64 (Intel):
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_darwin_amd64.tar.gz" | tar xz
# macOS arm64 (Apple Silicon):
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_darwin_arm64.tar.gz" | tar xz

# PATH の通ったディレクトリに移動（例: /usr/local/bin）
sudo mv trumpcards /usr/local/bin/
trumpcards --version
```

**Windows (PowerShell)**
```powershell
# VERSION には取得したいバージョン番号を指定（例: v3.12.0）
$VERSION = "v3.12.0"
$VER = $VERSION.TrimStart("v")

Invoke-WebRequest -Uri "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/$VERSION/trumpcards_${VER}_windows_amd64.zip" -OutFile "trumpcards.zip"
Expand-Archive -Path "trumpcards.zip" -DestinationPath "."
# trumpcards.exe を PATH の通ったディレクトリに移動してから実行
.\trumpcards.exe --version
```

#### ソースからビルド
```sh
git clone https://github.com/yuta-yoshinaga/go_trumpcards.git
cd go_trumpcards
```

### Run
```sh
go run ./cmd/trumpcards                      # インタラクティブモード (ゲーム選択・切り替え可能)
go run ./cmd/trumpcards --lang en            # インタラクティブモード (英語)
go run ./cmd/trumpcards blackjack            # ブラックジャック CLI
go run ./cmd/trumpcards --lang en blackjack  # ブラックジャック CLI (英語)
go run ./cmd/trumpcards poker      # 5枚ドローポーカー CLI
go run ./cmd/trumpcards oldmaid    # ババ抜き CLI
go run ./cmd/trumpcards daifugo    # 大富豪 CLI
go run ./cmd/trumpcards sevens     # 7並べ CLI
go run ./cmd/trumpcards doubt      # ダウト CLI
go run ./cmd/trumpcards holdem     # テキサスホールデム CLI
go run ./cmd/trumpcards omaha      # オマハホールデム CLI
go run ./cmd/trumpcards hearts     # ハーツ CLI
go run ./cmd/trumpcards memory     # 神経衰弱 CLI
go run ./cmd/trumpcards klondike   # クロンダイク CLI
go run ./cmd/trumpcards freecell   # フリーセル CLI
go run ./cmd/trumpcards baccarat   # バカラ CLI
go run ./cmd/trumpcards spades     # スペード CLI
go run ./cmd/trumpcards crazyeights # クレイジーエイト CLI
go run ./cmd/trumpcards ginrummy   # ジンラミー CLI
go run ./cmd/trumpcards spider     # スパイダーソリティア CLI
go run ./cmd/trumpcards napoleon   # ナポレオン CLI
go run ./cmd/trumpcards update     # 最新版にセルフアップデート
go run ./cmd/trumpcards web        # REST API + Web GUI サーバー起動 (CLI経由)
go run ./cmd/server                # REST API + Web GUI サーバー起動 (直接)
PORT=3000 go run ./cmd/trumpcards web  # カスタムポートで起動 (デフォルト: 8080)
PORT=3000 go run ./cmd/server          # カスタムポートで起動 (直接)
```

### Test
```sh
go test -tags test ./...  # 全テスト実行
```

### Documentation

📚 **[API Documentation](https://yuta-yoshinaga.github.io/go_trumpcards/)** — Go と TypeScript の自動生成 API ドキュメント、アーキテクチャガイド、リポジトリスナップショット

- **Go API Docs** — Domain, Use Case, Adapter, Infrastructure パッケージのドキュメント
- **TypeScript API Docs** — React コンポーネント、フック、ユーティリティ、API クライアントのドキュメント
- **Repomix Output** — AI コンテキスト用のリポジトリ圧縮スナップショット (デプロイの詳細は [GitHub Pages (Repomix)](#github-pages-repomix) を参照)

### Deploy
[render live](https://go-trumpcards.onrender.com/)
[render dev](https://go-trumpcards-dev.onrender.com/)

### GitHub Pages (Repomix)

When code is merged into the `master` branch, a GitHub Actions workflow automatically runs [repomix](https://github.com/yamadashy/repomix) and deploys the packed XML output to GitHub Pages at:

```
https://yuta-yoshinaga.github.io/go_trumpcards/repomix-output.txt
```

**One-time repository setup required:**
Go to **Settings > Pages** in the repository and set the "Build and deployment" Source to **GitHub Actions**.

### Docker
#### Build
```sh
docker build -t go_trumpcards .
```

#### Run
```sh
docker run --rm -d -p 8080:8080 go_trumpcards
docker run --rm -d -e PORT=3000 -p 3000:3000 go_trumpcards  # カスタムポート
```
Open [http://localhost:8080](http://localhost:8080) in your browser.

## Architecture
Clean Architectureを採用しています。依存の方向は外側から内側への一方向です。`golang-standards/project-layout` に準拠した `cmd/` + `internal/` 構成を採用しています。

```
cmd/
  trumpcards/         # CLIエントリーポイント（全ゲーム + Webサーバー）
  server/             # Webサーバー専用エントリーポイント
internal/
  domain/             # コアビジネスロジック（最内層）
  usecase/            # アプリケーションビジネスルール
    presenter/        # プレゼンターインターフェース
  adapter/
    controller/       # コマンドをユースケースにルーティング
    presenter/        # CUI/Web向けプレゼンター実装
  infrastructure/
    ui/               # CLIランナー
    web/              # REST APIサーバー (go-json-rest)
api/                  # OpenAPI仕様
frontend/             # Reactフロントエンドソース（Vite + React + TypeScript）
public/               # Webフロントエンドビルド済みアセット（Reactビルド出力 + 静的ファイル）
```

## Future Releases
その他のトランプカードも実装したい。

## Contribution
1. Fork it
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create new Pull Request

## License
[MIT](LICENSE)
