# Crazy Eights (クレイジーエイト) — Web GUIマニュアル

## アクセス

```sh
go run ./cmd/trumpcards web
# ブラウザで http://localhost:8080/#/crazyeights を開く
```

## ルール概要

- 4人プレイ（1人 + CPU×3）、52枚のトランプ
- 各プレイヤーに5枚ずつ配り、1枚を表向きにして捨て札の山を開始、残りは山札
- 捨て札の山のトップカードとスートまたはランクが一致するカードを出す
- **8はワイルド**: 出すと次のスートを選べる
- 出せるカードがなければ山札から1枚ドロー
- ラウンド終了: 誰かが手札を全て出し切ったら終了
- スコアリング: 8=50点、J/Q/K=10点、A=1点、その他=額面
- ポイント上限（デフォルト200点）に達したらマッチ終了

## Web GUI操作

### プレイフェーズ
- 手札のカードをクリックして選択（1枚）
- 「カードを出す」ボタンでプレイ
- 出せるカードがなければ「ドロー」ボタンで山札から1枚引く

### スート選択
- 8を出した後、スート選択ボタン（♠ / ♣ / ♥ / ♦）が表示される
- 選択したスートが次のプレイで要求される

### ラウンド終了
- 「次のラウンド」ボタンで進行

### 設定パネル
- **CPU難易度**: Easy / Normal / Hard
- **ポイント上限**: デフォルト200

### スコアテーブル
- 各プレイヤーのラウンドスコアと累積スコアを表示

## API

```
POST /crazyeights/exec
```

### リクエスト例

```json
{"command": "reset", "sessionId": "xxx"}
{"command": "play", "cardIndex": 2, "sessionId": "xxx"}
{"command": "draw", "sessionId": "xxx"}
{"command": "suit", "suit": 3, "sessionId": "xxx"}
{"command": "nextround", "sessionId": "xxx"}
{"command": "log", "sessionId": "xxx"}
```
