# go_trumpcards

トランプカードゲームアルゴリズムをGoで実装

[![Backend Coverage](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards/graph/badge.svg?flag=backend)](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards)
[![Frontend Coverage](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards/graph/badge.svg?flag=frontend)](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards)

## Vision

**世界中のあらゆるトランプゲームを、誰でも無料で遊べるようにする。**

このプロジェクトは、人間とAIコーディングエージェントが共にソフトウェアを創り上げる **「共創のリファレンスモデル」** です。AIエージェントが正確にコンテキストを理解し、高品質なコードを生成できる開発環境を整備することで、人間とAIの協調開発のベストプラクティスを示し続けます。

go_trumpcardsが目指す未来は、**あらゆる人がクリエイターとなり、自分が欲しいものを生成AIコーディングエージェントとの共創で実現できる世の中** です。

## Features

Go + Clean Architecture で実装した43種類のトランプゲーム。CLI と Web GUI（React + Go REST API）の2つのインターフェースで遊べます。Web GUI は日英多言語対応。

| ゲーム | コマンド | マニュアル |
|--------|----------|------------|
| ブラックジャック (BlackJack) | `blackjack` | [CUI](docs/manual/cui/blackjack.md) / [Web](docs/manual/web/blackjack.md) |
| ポーカー (5-card Draw) | `poker` | [CUI](docs/manual/cui/poker.md) / [Web](docs/manual/web/poker.md) |
| ババ抜き (Old Maid) | `oldmaid` | [CUI](docs/manual/cui/oldmaid.md) / [Web](docs/manual/web/oldmaid.md) |
| 大富豪 (Daifugo) | `daifugo` | [CUI](docs/manual/cui/daifugo.md) / [Web](docs/manual/web/daifugo.md) |
| 7並べ (Sevens) | `sevens` | [CUI](docs/manual/cui/sevens.md) / [Web](docs/manual/web/sevens.md) |
| ダウト (Doubt) | `doubt` | [CUI](docs/manual/cui/doubt.md) / [Web](docs/manual/web/doubt.md) |
| テキサスホールデム (Texas Hold'em) | `holdem` | [CUI](docs/manual/cui/holdem.md) / [Web](docs/manual/web/holdem.md) |
| オマハホールデム (Omaha Hold'em) | `omaha` | [CUI](docs/manual/cui/omaha.md) / [Web](docs/manual/web/omaha.md) |
| ショートデック (Short Deck / 6+ Hold'em) | `shortdeck` | [CUI](docs/manual/cui/shortdeck.md) / [Web](docs/manual/web/shortdeck.md) |
| パイナップルポーカー (Pineapple Poker) | `pineapple` | [CUI](docs/manual/cui/pineapple.md) / [Web](docs/manual/web/pineapple.md) |
| ハーツ (Hearts) | `hearts` | [CUI](docs/manual/cui/hearts.md) / [Web](docs/manual/web/hearts.md) |
| 神経衰弱 (Memory) | `memory` | [CUI](docs/manual/cui/memory.md) / [Web](docs/manual/web/memory.md) |
| クロンダイク (Klondike) | `klondike` | [CUI](docs/manual/cui/klondike.md) / [Web](docs/manual/web/klondike.md) |
| フリーセル (FreeCell) | `freecell` | [CUI](docs/manual/cui/freecell.md) / [Web](docs/manual/web/freecell.md) |
| バカラ (Baccarat) | `baccarat` | [CUI](docs/manual/cui/baccarat.md) / [Web](docs/manual/web/baccarat.md) |
| スペード (Spades) | `spades` | [CUI](docs/manual/cui/spades.md) / [Web](docs/manual/web/spades.md) |
| ツーテンジャック (Two Ten Jack) | `twotenjack` | [CUI](docs/manual/cui/twotenjack.md) / [Web](docs/manual/web/twotenjack.md) |
| クレイジーエイト (Crazy Eights) | `crazyeights` | [CUI](docs/manual/cui/crazyeights.md) / [Web](docs/manual/web/crazyeights.md) |
| ジンラミー (Gin Rummy) | `ginrummy` | [CUI](docs/manual/cui/ginrummy.md) / [Web](docs/manual/web/ginrummy.md) |
| カナスタ (Canasta) | `canasta` | [CUI](docs/manual/cui/canasta.md) / [Web](docs/manual/web/canasta.md) |
| スパイダーソリティア (Spider Solitaire) | `spider` | [CUI](docs/manual/cui/spider.md) / [Web](docs/manual/web/spider.md) |
| ナポレオン (Napoleon) | `napoleon` | [CUI](docs/manual/cui/napoleon.md) / [Web](docs/manual/web/napoleon.md) |
| インディアンポーカー (Indian Poker) | `indianpoker` | [CUI](docs/manual/cui/indianpoker.md) / [Web](docs/manual/web/indianpoker.md) |
| ビデオポーカー (Video Poker) | `videopoker` | [CUI](docs/manual/cui/videopoker.md) / [Web](docs/manual/web/videopoker.md) |
| デューシーズワイルド (Deuces Wild) | `deuceswild` | [CUI](docs/manual/cui/deuceswild.md) / [Web](docs/manual/web/deuceswild.md) |
| ジョーカーポーカー (Joker Poker) | `jokerpoker` | [CUI](docs/manual/cui/jokerpoker.md) / [Web](docs/manual/web/jokerpoker.md) |
| ユーカー (Euchre) | `euchre` | [CUI](docs/manual/cui/euchre.md) / [Web](docs/manual/web/euchre.md) |
| ピラミッド (Pyramid) | `pyramid` | [CUI](docs/manual/cui/pyramid.md) / [Web](docs/manual/web/pyramid.md) |
| トリピークス (TriPeaks) | `tripeaks` | [CUI](docs/manual/cui/tripeaks.md) / [Web](docs/manual/web/tripeaks.md) |
| クリベッジ (Cribbage) | `cribbage` | [CUI](docs/manual/cui/cribbage.md) / [Web](docs/manual/web/cribbage.md) |
| スリーカードポーカー (Three Card Poker) | `threecard` | [CUI](docs/manual/cui/threecard.md) / [Web](docs/manual/web/threecard.md) |
| オー・ヘル (Oh Hell) | `ohhell` | [CUI](docs/manual/cui/ohhell.md) / [Web](docs/manual/web/ohhell.md) |
| コントラクトブリッジ (Contract Bridge) | `bridge` | [CUI](docs/manual/cui/bridge.md) / [Web](docs/manual/web/bridge.md) |
| スピード (Speed) | `speed` | [CUI](docs/manual/cui/speed.md) / [Web](docs/manual/web/speed.md) |
| ゴーフィッシュ (Go Fish) | `gofish` | [CUI](docs/manual/cui/gofish.md) / [Web](docs/manual/web/gofish.md) |
| ピノクル (Pinochle) | `pinochle` | [CUI](docs/manual/cui/pinochle.md) / [Web](docs/manual/web/pinochle.md) |
| ゴルフ (Golf Solitaire) | `golf` | [CUI](docs/manual/cui/golf.md) / [Web](docs/manual/web/golf.md) |
| ぶたのしっぽ (Pig's Tail) | `pigtail` | [CUI](docs/manual/cui/pigtail.md) / [Web](docs/manual/web/pigtail.md) |
| セブンカード・スタッド (Seven Card Stud) | `sevencardstud` | [CUI](docs/manual/cui/sevencardstud.md) / [Web](docs/manual/web/sevencardstud.md) |
| クロックソリティア (Clock Solitaire) | `clocksolitaire` | [CUI](docs/manual/cui/clocksolitaire.md) / [Web](docs/manual/web/clocksolitaire.md) |
| ドゥラーク (Durak) | `durak` | [CUI](docs/manual/cui/durak.md) / [Web](docs/manual/web/durak.md) |
| フォーティシーブス (Forty Thieves) | `fortythieves` | [CUI](docs/manual/cui/fortythieves.md) / [Web](docs/manual/web/fortythieves.md) |
| パイゴウポーカー (Pai Gow Poker) | `paigow` | [CUI](docs/manual/cui/paigow.md) / [Web](docs/manual/web/paigow.md) |

## Demo

### Cloudflare (Edge)

- [Live (production)](https://go-trumpcards.pages.dev/)
- [Dev](https://go-trumpcards-staging.pages.dev/)

### Render (Docker)

- [Live (production)](https://go-trumpcards.onrender.com/)
- [Dev](https://go-trumpcards-dev.onrender.com/)
- [Swagger UI](https://go-trumpcards.onrender.com/swagger/)

## Getting Started

### Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.26.x |
| [Node.js](https://nodejs.org/) | 24.x |
| [Bun](https://bun.sh/) | 1.3.10 |

### Installation

#### go install

```sh
go install github.com/yuta-yoshinaga/go_trumpcards/cmd/trumpcards@latest
trumpcards  # or: trumpcards blackjack
```

#### GitHub Releases

Linux/macOS/Windows 向けのビルド済みバイナリは [GitHub Releases](https://github.com/yuta-yoshinaga/go_trumpcards/releases) から入手できます。

<details>
<summary>Linux / macOS</summary>

```sh
# 最新バージョンを https://github.com/yuta-yoshinaga/go_trumpcards/releases から取得して設定
VERSION=vX.Y.Z

# Linux amd64:
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_linux_amd64.tar.gz" | tar xz
# Linux arm64:
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_linux_arm64.tar.gz" | tar xz
# macOS amd64 (Intel):
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_darwin_amd64.tar.gz" | tar xz
# macOS arm64 (Apple Silicon):
curl -fsSL "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/${VERSION}/trumpcards_${VERSION#v}_darwin_arm64.tar.gz" | tar xz

sudo mv trumpcards /usr/local/bin/
trumpcards --version
```

</details>

<details>
<summary>Windows (PowerShell)</summary>

```powershell
# 最新バージョンを https://github.com/yuta-yoshinaga/go_trumpcards/releases から取得して設定
$VERSION = "vX.Y.Z"
$VER = $VERSION.TrimStart("v")

Invoke-WebRequest -Uri "https://github.com/yuta-yoshinaga/go_trumpcards/releases/download/$VERSION/trumpcards_${VER}_windows_amd64.zip" -OutFile "trumpcards.zip"
Expand-Archive -Path "trumpcards.zip" -DestinationPath "."
.\trumpcards.exe --version
```

</details>

#### Build from Source

```sh
git clone https://github.com/yuta-yoshinaga/go_trumpcards.git
cd go_trumpcards
go run ./cmd/trumpcards
```

### Usage

```sh
trumpcards                       # インタラクティブモード (ゲーム選択・切り替え可能)
trumpcards --lang en             # インタラクティブモード (英語)
trumpcards blackjack             # ブラックジャック CLI
trumpcards --lang en blackjack   # ブラックジャック CLI (英語)
trumpcards web                   # REST API + Web GUI サーバー起動
trumpcards web --port 3000       # カスタムポートで起動 (--port フラグ)
trumpcards update                # 最新版にセルフアップデート
PORT=3000 trumpcards web         # カスタムポートで起動 (環境変数)
source <(trumpcards completion bash)  # Bash 補完を有効化
```

### Docker

```sh
docker build -t go_trumpcards .
docker run --rm -d -p 8080:8080 go_trumpcards
# カスタムポート
docker run --rm -d -e PORT=3000 -p 3000:3000 go_trumpcards
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

## Development

### Test

```sh
go test -tags test ./...  # 全テスト実行
```

### Frontend

```sh
cd frontend && bun install   # 依存関係インストール
cd frontend && bun run build # ビルド
cd frontend && bun run check # Lint + フォーマットチェック
cd frontend && bun run test  # ユニットテスト
cd frontend && bun run e2e   # E2Eテスト
```

## Architecture

Clean Architecture を採用。依存の方向は外側から内側への一方向です。`golang-standards/project-layout` に準拠した `cmd/` + `internal/` 構成。

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
frontend/             # React フロントエンド（Vite + React + TypeScript）
public/               # ビルド済みアセット
```

詳細は [docs/architecture.md](docs/architecture.md) を参照。

## Documentation

- [API Documentation](https://yuta-yoshinaga.github.io/go_trumpcards/) — Go / TypeScript 自動生成ドキュメント
- [Repomix Output](https://yuta-yoshinaga.github.io/go_trumpcards/repomix-output.txt) — AIコンテキスト用リポジトリスナップショット（master マージ時に自動生成）
- [OpenAPI Spec](api/openapi.yaml)
- [Architecture](docs/architecture.md)
- [Backend Design (UML)](docs/design/backend.md) — クラス図・シーケンス図・状態遷移図
- [Frontend Design (UML)](docs/design/frontend.md) — コンポーネント図・シーケンス図・状態遷移図
- [Game Descriptions](docs/games.md)
- [ADR (Architecture Decision Records)](docs/adr/)

## Contributing

コントリビューション歓迎です！詳細は [CONTRIBUTING.md](CONTRIBUTING.md) を参照してください。

1. Fork it
2. Create your feature branch (`git checkout -b feat/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feat/amazing-feature`)
5. Create new Pull Request

## License

[MIT](LICENSE) © 2020 Yuta Yoshinaga
