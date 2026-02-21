# go_trumpcards
トランプカードゲームアルゴリズムをGoで実装

[![Backend Coverage](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards/graph/badge.svg?flag=backend)](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards)
[![Frontend Coverage](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards/graph/badge.svg?flag=frontend)](https://codecov.io/gh/yuta-yoshinaga/go_trumpcards)

## Description
トランプカードのアルゴリズムをGo+Clean Architectureで実装したプロジェクトです。

以下のゲームを実装しています：

- **ブラックジャック (BlackJack)**: CLI および Web GUIで遊べます
- **ポーカー (5-card Draw Poker)**: CLIで遊べます
- **ババ抜き (Old Maid)**: CLI および Web GUIで遊べます
- **大富豪 (Daifugo)**: CLI および Web GUIで遊べます
- **7並べ (Sevens)**: CLI および Web GUIで遊べます

## Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/) | 1.24.x |
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
go run main.go blackjack  # ブラックジャック CLI
go run main.go poker    # 5枚ドローポーカー CLI
go run main.go oldmaid  # ババ抜き CLI
go run main.go daifugo  # 大富豪 CLI
go run main.go sevens   # 7並べ CLI
go run main.go web      # REST API + Web GUI サーバー起動
```

### Test
```sh
go test ./...  # 全テスト実行
```

### Deploy
[render live](https://go-trumpcards.onrender.com/)
[render dev](https://go-trumpcards-dev.onrender.com/)

### Docker
#### Build
```sh
docker build -t go_trumpcards .
```

#### Run
```sh
docker run -p 8080:8080 go_trumpcards
```
Open [http://localhost:8080](http://localhost:8080) in your browser.

## Architecture
Clean Architectureを採用しています。依存の方向は外側から内側への一方向です。

```
entities/             # コアビジネスロジック（最内層）
usecases/             # アプリケーションビジネスルール
interface_adapters/   # レイヤー間のデータ変換
frameworks_drivers/   # 最外層（CLI・Webサーバー）
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
