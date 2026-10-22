# ADR-0045: 合法手をサーバデータとして返す

## Status

Accepted (Hearts #11917 と Euchre で同じ形のまま移行できた)

## Date

2026-10-19

更新: 2026-10-20

## Context

Issue #11913。フロントがルールを再実装した client-side mirror は 31 本 (pages 8 / utils 16 / utils/hints 7)。一方 `validPlays` / `validPlayIndices` をサーバから受け取る型が既に 18 本ある。mirror に意図 (往復なしの即時フィードバック) の記録はない。合法手は状態と同じ応答に乗るので、データ化しても往復は増えない。Hearts の `heartsLegal.ts` は「合法カードを強調する」ためだけに作られ、`validatePlay` (`internal/domain/Hearts.go`) の写しだった。

代替案:

- (a) mirror を残し Go domain と突き合わせる試験で縛る — 2 実装が残り、規則変更のたびに両方直す。
- (b) presenter が合法手をデータで返し mirror を削除する — 採用。

## Decision

合法手は domain の検証関数 (`validatePlay` 等) を再利用した `GetValidPlayIndices` から presenter が `validPlays` として返す。フロントは受け取った添字を表示に使うだけで、規則を持たない。移行は 1 ゲーム 1 PR、触るついで (ADR-0044 と同じ方針)。WASM の gzip 残容量を確認してから着手する (extra7 は 2026-10-18 実測 852,512 B / 1 MB)。

### 改訂 (2026-10-20): ソリティアの移動先

手札の添字の列では表せない「選んだ札をどこに置けるか」も、同じ方針でサーバのデータにする。ユーザ承認済み (#11913 の改訂案)。

- 形は Bristol (#4813) で既に使っているものに揃える。通常の状態応答に、移動元ごとの移動先を載せる: `legalTargets: Record<string, { tableau: number[]; foundation: number[] }>`。移動先の無い移動元はキーごと省く。
- キーは `"<zone>-<col>"`。埋もれた札から並びごと動かせるゲームでは `"<zone>-<col>-<cardIndex>"` にする (Perseverance)。
- 判定は domain の `canPlaceOn*` などをそのまま通す。presenter には規則を書かない。
- #5581 で Baker's Dozen と Perseverance がこの形を見送った理由は「CUI 専用の応答が Web の経路に紛れ込む」ことだった (`targets` コマンドの応答を Web に使う案だった)。この改訂では `targets` の経路は使わず、通常の盤面の応答に載せるので、その心配は当たらない。
- 試しに移すのは Baker's Dozen と Perseverance。残りは触るついでに 1 ゲーム 1 PR で移す。MrsMop のように移動元の形が違うゲームは、そのときにキーの形を決める。

## Consequences

規則の SSoT が Go domain に一本化される。mirror 削除ごとにフロントのテストも消える。CLI/CUI も同じ関数を使える。

Euchre は `GetValidPlayIndices` が既にあったため domain の変更は不要で、presenter の手番判定とレスポンス field の追加だけで移行できた。Belote、Omi、Wizard、Coinche、Bezique も同じく domain 変更なしで移行した。Bezique は終盤 (山札切れ) で後手のときだけリングを出す表示条件をフロントに残している。`isEndgame` はサーバの値なので規則の写しではない。

CrazyEights と Prsi は domain に `GetValidPlayIndices` があったが interface に出ていなかったため、interface と手書きの mock に足して同じ形で移行した。Prsi は hook の `canPlay` が「出す」を止めているので、その判定も `validPlays` から決める。Prsi のペナルティ中の読み上げ文はサーバの `penaltyDrawCount` を見る表示条件として残している。

Batak と Call Break は「なぜ出せないか」のツールチップを写しで計算していた。domain に `GetPlayRejectCodes` を足し、`validatePlay` が返すエラーのメッセージコードを手札と同じ並びで `playRejectCodes` として返す (出せる札は空文字)。ページはコードを既存のツールチップの文言に対応させるだけで、知らないコードは汎用のツールチップに任せる。合法手そのもの (`validPlayIndices`) の返し方は変えていない。

### ソリティアの移動先 (2026-10-20〜22)

Decision の改訂 (2026-10-20) の形で、Baker's Dozen と Perseverance から移した。Perseverance は domain に、埋もれた札から始まる並びの移動先を返す `LegalTargetsFrom(col, cardIndex)` を足し、キーを `tableau-<col>-<cardIndex>` にした。Beleaguered Castle、Citadel、Fortress、Somerset、Streets and Alleys は domain に `LegalTargets(fromCol)` を足して移行した。組札の候補は各ページの従来の表示に合わせた。castle 系の 4 本は受け入れる組札をすべて返し、Streets and Alleys は `findFoundation` が返す最初の 1 山だけを返す。Mrs. Mop は domain の既存の `LegalTargets(fromCol, cardIndex)` を presenter から呼ぶだけで移行した (`foundation` は常に空)。Perseverance は掴める札も `runStarts` で返し、`perseveranceStartsRun.ts` を削除した。King Albert、Flower Garden、Russian Solitaire、Miss Milligan、Rank and File も移行した (2026-10-21)。King Albert と Flower Garden は予備の札置き場を `reserve-<idx>` のキーで返す。Rank and File は捨て札を `waste`、Miss Milligan は保留中の札を `waived` のキーで返す。Russian Solitaire は表向きの札ならどれでも動かせるので、札ごとに 3 部分のキーで返す。組札の候補は各ページの従来の表示に合わせた。King Albert、Flower Garden、Rank and File は最初の 1 山、Miss Milligan は受け入れる組札すべて、Russian Solitaire はスートで決まる 1 山。Miss Milligan は `missMilliganRuns.ts` (2026-10-20 の数え上げから漏れていた) の代わりに `runStarts` も返す。TriPeaks は、上札に乗せられる表の札を `playableCells: ["<row>-<col>"]` で返す (ユーザ承認済み、2026-10-21)。CUI の presenter が同じ判定を書き直していた `triPeaksAdjacentRank` も、domain の `PlayableCells()` に置き換えた。

組札ショートカットの写し (`*FoundationTarget.ts`) も同じ `legalTargets` で置き換えた (2026-10-21)。FreeCell、Baker's Game (FreeCell と応答を共有)、Eight Off、Penguin、Stalactites、Easthaven、Forty Thieves、Forty and Eight。ショートカットは `legalTargets[key].foundation[0]` へ送るだけで、規則を持たない。組札の候補は各ページの従来の挙動に合わせた。Forty and Eight だけは受け入れる組札と列をすべて返し、Stalactites は `foundationIndexFor` の 1 山、他はスートで決まる 1 山か `findFoundation` の最初の 1 山。新しい応答の形 (ユーザ承認済み) として、Shamrocks は `fan-<idx>` のキーに `{ fan, foundation }`、Crazy Quilt は `quilt-<idx>` / `waste` のキーに `{ foundation, waste }` を返す。Napoleon's Square、Congress、Duchess、American Toad も移行した (2026-10-22)。移動元ごとの空き列の規則は domain の関数がそのまま表す。Congress の山札は空の山だけに置け、山札が尽きたら置き先を返さない。Duchess は基準ランクが決まるまで何も返さず、予備札だけは空き列に置ける。American Toad はタブローから空き列へは動かせない。Duchess と American Toad では、並びの先頭でない札を選んだときに移動先を出さず、並びの途中の先頭では組札を出さないようにした (以前の写しはこの 2 点を見ていなかった)。Sultan は場札の列が無いので `legalTargets: Record<"divan-<idx>" | "waste", { foundation: number[] }>` で返す (ユーザ承認済み)。Diplomat は既存の `tableauMoveTargets` を `legalTargets` に置き換えて削除し、組札の「次のランク」の表示用に `foundationNextRank` を足した (ユーザ承認済み)。

### 数え直し (2026-10-22)

2026-10-20 の数え上げは、`*LegalTargets.ts` などのファイル名で探したため漏れがあった。`frontend/src/utils` と `pages` を、規則の写しを示すコメント (`Sync:` や domain の関数名) とファイル名で数え直した。`utils/hints` (助言) と `utils/cli` は対象外。

既に承認済みの形で移せる 13 本は同じ日に移した (2026-10-22)。

- `legalTargets`:
  - Crescent、Salic Law、Double Klondike、Nertz は受け入れる組札をすべて返す。Double Klondike の組札は一番上の札で判定する。Nertz は自分の Nertz 山を `nertz` のキーで返し、人間のプレイヤーの分だけを載せる。
  - Agnes はスートで決まる 1 山を返す。
  - St. Helena は、どの列に届くかの制約も移動先の判定に含めて返す。
  - Scorpion は空き列に置けるキングの移動先も返すが、ページの強調表示は一番下の札にしか付かないので、空き列の分は表示に使われない。
- `runStarts`: Curds and Whey、Simple Simon、Spiderette、Spider、Mrs. Mop。`spiderMovableRun` は削除した。Spiderette の golden fixture は Go 側の試験 (`Spiderette_golden_test.go`) がまだ読むので残している。
- `playRejectCodes`: Klaverjas。Batak / Call Break と同じ形で返す。

残った写しのうち 6 本は、新しい応答の形の承認を得て同じ日に移した (2026-10-22)。

- 自動移動の優先順位: Simple Simon と Curds and Whey は、run の先頭ごとに `legalTargets` (`tableau-<col>-<cardIndex>`、`foundation` は常に空) を返す。`tableau` は優先順に並んでいて、ページは `tableau[0]` へ送るだけ。順番は domain の `AutoMoveTargets` が決める。同じスートで 1 つ上の札、次に Simple Simon はスートを問わず 1 つ上の札 (Curds and Whey は同じランクの札)、最後に空き列。空き列は run の上に札があるときだけ返す。
- Bisley は `legalTargets: Record<"tableau-<col>", { tableau: number[] }>` を返す (組札の配列は持たない)。domain の移動先の判定を `canStackOn` に括り出し、移動・`LegalTargets`・ヒントが同じ判定を通る。一番上の札のツールチップは「受け付ける札」から「動かせる列」の表示に変えた (ユーザ承認済み)。
- Fourteen Out は、既に応答に載っていた `removablePairs` を使い、選んだ列と組める列を `legalTargets` (`tableau-<col>`) で返す。ヒントも `legalTargets` を読むようにして、写しを削除した。
- Monte Carlo は `removablePairs: number` (Fourteen Out と同じ名前と型)、Gaps は行ごとの `lockedPrefixLengths: number[]` を返す。どちらも CUI が既に使っている domain の関数をそのまま返す。

役や組み合わせの検証 (`machiavelliRearrange.ts`、`doudizhuComboValidator.ts`、`burracoMeld.ts`、`tonkMeldIndices.ts`、`tongitsMeldIndices.ts`) はこの ADR の対象外とする (ユーザ承認済み)。どれもプレイヤーが組んだ選択を検証するもので、選ぶ前にはサーバに分からない。サーバに判定させると選ぶたびに往復が要り、合法手を状態と同じ応答に載せるというこの ADR の前提に合わない。`panMeldCandidates.ts` はヒントの設定で出る表示なので、`utils/hints` と同じ扱いで対象外。これで、この ADR の対象になる合法性の写しは残っていない。

点数・順位・表示用の計算 (`chinesePokerFoul.ts`、`heartsLimit.ts` など) は合法手ではないので、この ADR の対象外。

写しではないもの: `schnapsenLegal.ts` と `michiganPlayable.ts` はサーバから受け取った値の表示条件だけを持つ。

スコープ外: `utils/hints` の約 370 本の戦略ヒント (どれを出すべきか) は規則ではなく助言なので対象外。対象は合法性の写しだけ。
