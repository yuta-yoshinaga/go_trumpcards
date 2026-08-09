# ピケ（CUI版）遊び方

## ゲーム概要

Piquetは2人用の古典的トリックテイキング。32枚デッキ（7〜A）を使い、交換・宣言・12トリックのプレイを6ディール繰り返してパルティ通算で勝敗を決める。

## 起動方法

```sh
go run ./cmd/trumpcards piquet
go run ./cmd/trumpcards --lang en piquet  # 英語モード
```

## ルール

Webマニュアル ([../web/piquet.md](../web/piquet.md)) と同じ。簡潔に：

- Point / Sequence / Set の3宣言。
- Carte blanche +10、Repique +60、Pique +30、Cards +10、Capot +40。
- 6ディールのパルティ。ルビコン: 敗者 < 100 → 勝者 += 100 + W + L。

## コマンド

| コマンド | 説明 |
|----------|------|
| `e <i,j,k>` | Elder交換 (1..5枚) |
| `y <i,j,k>` | Younger交換 (0..3枚)。`y` 単独で0枚交換 (パス) |
| `d`, `declare` | 次の宣言を1ステップ進める |
| `p <i>` | 手札 i 番目をプレイ |
| `nd`, `nextdeal` | 次ディールへ |
| `h`, `hint` | ヒント |
| `l`, `log` | アクションログ |
| `r`, `reset` | リセット |
| `q`, `quit` | 終了 |

## ゲームの流れ

```mermaid
stateDiagram-v2
    [*] --> Exchange
    Exchange --> Declaration: Younger 交換完了
    Declaration --> Play: 3宣言完了
    Play --> Score: 12トリック完了
    Score --> Exchange: NextDeal (deal < 6)
    Score --> GameEnd: NextDeal (deal == 6)
    GameEnd --> [*]
```
