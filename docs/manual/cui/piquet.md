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

## ゲームの流れ

```mermaid
flowchart TD
    A["交換: Elder が上位5枚から1〜5枚 → Younger が下位3枚から0〜3枚"] --> B["宣言: Point → Sequence → Set を自動比較"]
    B --> C["12トリックをプレイ"]
    C --> D["ディール集計（cards +10 / capot +40）"]
    D --> E{"6ディール目?"}
    E -- いいえ --> A
    E -- はい --> F["パルティ判定（ルビコン含む）・ゲーム終了"]
```

## コマンド一覧

| コマンド | 短縮形 | 説明 |
|----------|--------|------|
| `elder <i,j,k>` | `e` | Elder交換 (1..5枚) |
| `younger <i,j,k>` | `y` | Younger交換 (0..3枚)。`y` 単独で0枚交換 (パス) |
| `declare` | `d` | 次の宣言を1ステップ進める |
| `play <i>` | `p` | 手札 i 番目をプレイ |
| `nextdeal` | `nd` | 次ディールへ |
| `hint` | `h` | ヒント |
| `log` | `l` | アクションログ |
| `reset` | `r` | リセット |
| `quit` | `q` | 終了 |
| `help` | `?` | コマンド一覧を表示 |
