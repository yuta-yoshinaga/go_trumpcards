# go_trumpcards
トランプカードゲームアルゴリズムをGoで実装

[![Backend Coverage](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards/graph/badge.svg?flag=backend)](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards)
[![Frontend Coverage](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards/graph/badge.svg?flag=frontend)](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards)

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
- **ハーツ (Hearts)**: CLI および Web GUIで遊べます（1人 vs CPU×3、トリックテイキング、シュート・ザ・ムーン、3段階CPU難易度） — [CUI版マニュアル](docs/manual/cui/hearts.md) / [Web版マニュアル](docs/manual/web/hearts.md)
- **神経衰弱 (Memory)**: CLI および Web GUIで遊べます（1人 vs CPU×3、52枚のカードから同ランクのペアを揃える、3段階CPU難易度） — [CUI版マニュアル](docs/manual/cui/memory.md) / [Web版マニュアル](docs/manual/web/memory.md)
- **クロンダイク (Klondike)**: CLI および Web GUIで遊べます（1人用ソリティア、52枚のカード、7列のタブロー、山札・ウェスト・4つの組札） — [CUI版マニュアル](docs/manual/cui/klondike.md) / [Web版マニュアル](docs/manual/web/klondike.md)
- **バカラ (Baccarat)**: CLI および Web GUIで遊べます（プレイヤー/バンカー/タイベット、第3カードルール、5%バンカーコミッション、チップシステム） — [CUI版マニュアル](docs/manual/cui/baccarat.md) / [Web版マニュアル](docs/manual/web/baccarat.md)

## Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [npm](https://www.npmjs.com/) | 11.x |

## Usage
### Install
```sh
git clone https://github.com/yuta-yoshinaga/go_trumpcards.git
cd go_trumpcards
```

### Run
```sh
go run ./cmd/cli                      # インタラクティブモード (ゲーム選択・切り替え可能)
go run ./cmd/cli --lang en            # インタラクティブモード (英語)
go run ./cmd/cli blackjack            # ブラックジャック CLI
go run ./cmd/cli --lang en blackjack  # ブラックジャック CLI (英語)
go run ./cmd/cli poker      # 5枚ドローポーカー CLI
go run ./cmd/cli oldmaid    # ババ抜き CLI
go run ./cmd/cli daifugo    # 大富豪 CLI
go run ./cmd/cli sevens     # 7並べ CLI
go run ./cmd/cli doubt      # ダウト CLI
go run ./cmd/cli holdem     # テキサスホールデム CLI
go run ./cmd/cli hearts     # ハーツ CLI
go run ./cmd/cli memory     # 神経衰弱 CLI
go run ./cmd/cli klondike   # クロンダイク CLI
go run ./cmd/cli baccarat   # バカラ CLI
go run ./cmd/cli web        # REST API + Web GUI サーバー起動 (CLI経由)
go run ./cmd/server         # REST API + Web GUI サーバー起動 (直接)
```

### Test
```sh
go test ./...  # 全テスト実行
```

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
```
Open [http://localhost:8080](http://localhost:8080) in your browser.

## Architecture
Clean Architectureを採用しています。依存の方向は外側から内側への一方向です。`golang-standards/project-layout` に準拠した `cmd/` + `internal/` 構成を採用しています。

```
cmd/
  cli/                # CLIエントリーポイント（全ゲーム + Webサーバー）
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
