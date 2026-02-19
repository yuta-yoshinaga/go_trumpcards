# go_trumpcards
トランプカードゲームアルゴリズムをGoで実装

## Description
トランプカードのアルゴリズムをGo+Clean Architectureで実装したプロジェクトです。

以下のゲームを実装しています：

- **ブラックジャック (BlackJack)**: CLI および Web GUIで遊べます
- **ポーカー (5-card Draw Poker)**: CLIで遊べます
- **ババ抜き (Old Maid)**: CLI および Web GUIで遊べます
- **大富豪 (Daifugo)**: Web GUIで遊べます

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
go run main.go web      # REST API + Web GUI サーバー起動
```

### Test
```sh
go test ./...  # 全テスト実行
```

### Deploy
[render live](https://go-trumpcards.onrender.com/)
[render dev](https://go-trumpcards-dev.onrender.com/)

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
