# ADR-0040: 棋譜の説明文を DetailCode + DetailParams で持ち、文面は presenter で組む

## Status

Proposed

## Date

2026-09-13

## Context

`domain.ActionLogEntry.Detail` は自由形式の文字列で、いま 3 つの役割を掛け持ちしている。

1. **棋譜の本文。** `internal/adapter/presenter/action_log_helper.go:142` が
   `T%d [%s] %s: %s` で組み、CUI の棋譜表示と Web の `ActionLogPanel`
   (`frontend/src/components/ActionLogPanel.tsx:16` が同じ書式で組み直している) が読む。
2. **presenter が数値を取り出す元。** Rams が `Detail` を日本語の文面ごとパースしている
   (`RamsCuiPresenter.go:158` は `strings.Fields` + `Atoi`、`RamsWebPresenter.go:136,139` は
   `fmt.Sscanf(entry.Detail, "0 トリックで %d 支払い", ...)`)。同じ 1 行から同じ数を取るのに
   CUI は「最後の数字」、Web は「この日本語の型に一致する数字」を見る。
3. **セッション状態の一部。** `json:"al"` は `internal/domain` に **365 箇所**あり、
   そのぶんのゲームが `ActionLog []*ActionLogEntry` を永続化する。KV セッションは 1 手ごとに伸び、本番の Klondike で 117 手目に
   503 / Cloudflare 1102 を再現した実績がある。

この掛け持ちが実際に壊した例:

- **ロケール漏れ (#7179 / #7180 / #7181 / #7182)。** `Detail` の文面は書いた人の言語で
  固定される。実測すると `internal/domain` の非テストファイルにある `appendLog` 系呼び出し **2,604 件**
  (371 ファイル) のうち、Detail に文字列リテラルを渡しているものが 2,423 件。その内訳は
  日本語リテラル **513 件 (105 ファイル)**、英語リテラル **1,910 件 (249 ファイル)** で、
  **25 ファイルは同じゲームの中で両方**を使っている。どのロケールを選んでも棋譜の
  半分以上が読めない。
- **文面を直すと黙って壊れる (#7767)。** Rams の `Sscanf` は不一致を握り潰していて
  (`_, _ =`)、0 が入るだけ。「`Detail` は日本語である」ことが暗黙の契約になっており、
  英語化した瞬間に Web の内訳が全部 0 になる。
- **`Detail` を空にすると別の面が壊れる。** 2026-09-13、Piquet で画面メッセージを i18n
  キー経由に移した際に `appendLog` の Detail を空文字にし、棋譜が
  `T5 [Elder] trick_win: ` になる退行を作りかけた。棋譜側のテストが `if out == ""`
  しか見ていなかったので気づけなかった。

検討した選択肢は次のとおりである。

1. **`Detail` を i18n キーそのものにする。** `appendLog(idx, "trick_win", "piquet.trickWin", cards)`
   のようにキーを入れ、presenter が `i18n.T` に通す。1 フィールドで済み、ワイヤも増えない。
   ただし **パラメータを持てない。** 実測 2,604 件のうち数値や名前を埋め込むもの
   (`fmt.Sprintf` を通しているもの) が大半で、`"%s wins trick %d"` を 1 キーには落とせない。
   キーに値を埋め込む (`"piquet.trickWin.3"`) と翻訳ファイルが組合せ爆発する。却下。
2. **`Detail` を残したまま `DetailCode` / `DetailParams` を足す。** 移行が段階的にでき、
   既存の 2,695 件をそのままにできる。ただし **3 者が併存する期間が終わらない** —
   「`Detail` を空にしてよいか」が呼び出しごとに変わり、今回の Piquet の退行がそのまま
   再現する。さらに 365 箇所が永続化する `al` にフィールドが 2 つ増え、
   1 手ごとに伸びるセッションがさらに太る。
3. **`Detail` を `DetailCode` + `DetailParams` に置き換え、人間が読む文面は presenter が
   ロケールから組む。** ワイヤの `d` を `dc` (コード) と `dp` (パラメータ) に差し替える。

### 数え方 (再現用)

上の数は `internal/domain/*.go` から `_test.go` を除き、`appendLog` / `appendLogAt` /
`AppendActionLog` / `addActionLog` の**呼び出し**のみを数えたもの (ヘルパ自身の
`func` 宣言は除外)。言語の判定は同じ行の文字列リテラルに仮名・漢字が含まれるかで行った。
`json:"al"` は `grep -rn 'json:"al"' internal/` の件数。
**最初の版はテストファイルを含めて 2,695 / 「`"al"` を含むファイル数」で 427 と書いていた。
どちらも数え方が誤りで、レビューの指摘で測り直した。**

## Decision

3 を採用する。

```go
type ActionLogEntry struct {
    TurnNumber   int
    PlayerIdx    int
    ActionType   string
    DetailCode   string            // 例: "piquet.log.trickWin"
    DetailParams map[string]string // 例: {"trick": "3"}
    Cards        []*Card
}
```

- **ワイヤは `d` を落として `dc` / `dp` にする。** `dp` は空なら省略する
  (`json:"dp,omitempty"`)。実測で params を持たない呼び出しが相当数あるため、
  多くのエントリはむしろ現状より短くなる。選択肢 2 と違い**フィールドは増えない**ので、
  KV セッションの伸びを悪化させない。
- **文面は presenter が組む。** `action_log_helper.go` が `i18n.Tf(DetailCode, DetailParams...)`
  を通してから `T%d [%s] %s: %s` に流し込む。Web は `dc`/`dp` をそのまま受け取り、
  `ActionLogPanel` が `react-i18next` で組む。**同じ棋譜がロケールごとに正しく出る**のは
  ここで初めて成立する。
- **presenter が値を取り出すときは `DetailParams` を読む。** Rams の
  `lastActionLogNumber` と 2 本の `Sscanf` は削除する。文字列解析はどこにも残さない。
- **`DetailCode` が空のエントリは、文面のない行として描く。** 空文字を
  `i18n.T` に通してキー名を出さないこと。今回の Piquet のように「コードを入れ忘れた」
  変更が `T5 [Elder] trick_win: ` を出すのではなく、テストで落ちるようにする。

### 移行

2,604 呼び出しを一度に変えることはしない。ゲーム単位で移す。

1. `ActionLogEntry` を新しい形にし、`appendLog` の第 3 引数の型を変える。
   コンパイルエラーが移行対象の完全な一覧になる (`Detail string` を消すので、
   移し忘れたゲームはビルドが通らない)。**「タグ無しの静かな共存」を作らない**のが
   この順序の目的である。
2. 各ゲームの日本語/英語リテラルを `internal/i18n/locales/{ja,en}/<game>.json` の
   `log.*` に移す。`check-message-codes.mjs` と locale-parity ガードが両ロケールの
   存在を強制する。
3. `action_log_helper_test.go` に「`DetailCode` が空のエントリでキー名が画面に出ない」
   否定対照を置く。

### 保存済みセッションの扱い

`al` の要素の形が変わるため、**旧形式 (`d`) を読む `UnmarshalJSON` の後方互換分岐を残す。**
旧エントリは `DetailCode` を空、`Detail` 相当の文字列を `DetailParams{"legacy": …}` に
入れて描く。KV セッションは有限寿命なので、この分岐は 1 リリース分だけ置き、
次のリリースで削除する。

## Consequences

**良くなること**

- #7179 / #7180 / #7181 / #7182 が「そのゲームの `log.*` を訳す」だけの作業になる。
  4 層をまたぐ設計判断は本 ADR に閉じる。
- #7767 の 3 通りの文字列解析が消え、CUI と Web が同じ数を見ることが型で保証される。
- 文面を直しても数値の取り出しが壊れない。翻訳が壊れてもコードは動く。
- ワイヤはむしろ縮む可能性が高い (params 無しのエントリで `dc` は `d` より短いことが多い)。

**悪くなること・引き受けるコスト**

- **2,604 呼び出しの書き換え**が必要で、ゲームあたり数件〜数十件になる。
  ステップ 1 でビルドを壊してから移すため、**移行期間中は develop をマージできない。**
  ゲーム単位ではなく、`appendLog` のシグネチャ変更とその全呼び出しの書き換えを
  **1 つの PR に閉じる**必要がある。これは本 repo で最大級の機械的変更になる。
- ロケールファイルに `log.*` が 2,423 件分 (重複を除いても数百件) 増える。
  TinyGo Worker のバイナリに載る i18n JSON が太るので、**8 worker の gzip 余裕を
  移行前に実測する**こと (`.claude/skills/rebucket-game/scripts/measure.sh`)。
- 旧形式の後方互換分岐を 1 リリース抱える。

**やらないこと**

- `ActionType` は触らない。あれは既に機械可読で、presenter の分岐が読んでいる。
- 棋譜の書式 (`T%d [%s] %s: %s`) は変えない。変わるのは最後の `%s` の作り方だけ。
