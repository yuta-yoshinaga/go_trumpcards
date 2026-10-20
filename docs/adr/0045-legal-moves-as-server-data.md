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

## Consequences

規則の SSoT が Go domain に一本化される。mirror 削除ごとにフロントのテストも消える。CLI/CUI も同じ関数を使える。

Euchre は `GetValidPlayIndices` が既にあったため domain の変更は不要で、presenter の手番判定とレスポンス field の追加だけで移行できた。Belote、Omi、Wizard、Coinche、Bezique も同じく domain 変更なしで移行した。Bezique は終盤 (山札切れ) で後手のときだけリングを出す表示条件をフロントに残している。`isEndgame` はサーバの値なので規則の写しではない。

`frontend/src/utils` に残る合法性まわりの写し (2026-10-20 に実物を数えた):

- `validPlays` の形で移せるが interface と mock の変更が要る: CrazyEights、Prsi (2 本)。
- 合法かどうかではなく「なぜ出せないか」の理由の写し: Batak、Call Break (2 本)。どちらも合法手は既にサーバから受け取っている。
- ソリティアの移動先・動かせる札の計算: `*LegalTargets.ts` 18 本と `shamrocksLegalMove.ts`、`triPeaksPlayable.ts` (20 本)。どれも選択中の札や盤面に対する計算で、手札の添字の列である `validPlays` では置き換えられない。移すなら別の応答の設計 (この ADR の改訂) が要る。
- 写しではないもの: `schnapsenLegal.ts` と `michiganPlayable.ts` はサーバから受け取った値の表示条件だけを持つ。

スコープ外: `utils/hints` の約 370 本の戦略ヒント (どれを出すべきか) は規則ではなく助言なので対象外。対象は合法性の写しだけ。
