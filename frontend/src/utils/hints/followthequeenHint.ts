import type { Card, FollowTheQueenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { FollowTheQueenPhase } from '../../types/phases';

/** 賭けのあるストリート。ここ以外では助言しない。 */
const BETTING_STREETS: readonly number[] = [
  FollowTheQueenPhase.THIRD_STREET,
  FollowTheQueenPhase.FOURTH_STREET,
  FollowTheQueenPhase.FIFTH_STREET,
  FollowTheQueenPhase.SIXTH_STREET,
  FollowTheQueenPhase.SEVENTH_STREET,
];

/** 10 以上を「高い札」とみなす。A は 1 で届く。 */
const HIGH_CARD_FROM = 10;
const ACE = 1;

/**
 * Returns a frontend {@link HintResult} for Follow the Queen, or null when no
 * suggestion is available.
 *
 * **`handRank` is not read**, for the reason found in review on #4622: the
 * presenter fills it at showdown, so a made-hand branch keyed on it is dead on a
 * betting street while unit tests that set the field go on passing. The hand is
 * judged from the cards the human can see instead.
 *
 * What this game adds over Five Card Stud is the **bring-in**. On third street
 * the lowest door card is forced to open (`determineBringIn`,
 * `internal/domain/FollowTheQueen.go:328`), and being that seat is not a reason
 * to like your hand — it is the opposite. The hint says so rather than reading
 * the forced bet as strength.
 *
 */
export function getFollowTheQueenHint(state: FollowTheQueenResponse): HintResult | null {
  if (state.gameEndFlag || !BETTING_STREETS.includes(state.phase)) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.folded || human.allIn || state.currentTurn !== human.id) return null;

  const cards = [...human.holeCards, ...human.doorCards];
  if (cards.length === 0) return null;

  // **ブリングインは強さではない。**一番低い門札だから出させられているだけ。
  if (state.phase === FollowTheQueenPhase.THIRD_STREET && state.bringInPlayerIdx === human.id) {
    return { targetAction: 'call', reason: 'frontendHint.followthequeenBringIn', confidence: 'moderate' };
  }

  // 既に払い込んでいる分は差し引く。同額まで出していれば負債はない。
  const owed = Math.max(0, state.lastBet - human.currentBet);

  // **ワイルドの枚数が他のどの条件より大きい。**ワイルド1枚は実質ペア以上、
  // 2枚ならスリーカード以上が確定する。ペア判定を先に置くと、同じ手を
  // 「ワンペア」と言ってしまい強さを2段階見誤る。
  const wilds = cards.filter((c) => isWild(c, state.wildRank)).length;
  if (wilds >= 2) {
    return { targetAction: 'raise', reason: 'frontendHint.followthequeenRaiseWildTwo', confidence: 'strong' };
  }
  if (wilds === 1) {
    return { targetAction: 'raise', reason: 'frontendHint.followthequeenRaiseWildOne', confidence: 'moderate' };
  }

  if (hasPair(cards)) {
    return { targetAction: 'raise', reason: 'frontendHint.followthequeenRaisePair', confidence: 'moderate' };
  }

  // **払う額が無ければ「コール」ボタンは無い。**`lastBet` と `currentBet` は
  // ストリートごとに 0 に戻る (`advancePhase`, FollowTheQueen.go:504) ので、
  // 5th street 以降で human が先に動く場面は普通にある。そのとき
  // `BettingControls` が出すのは Bet と Check だけで、Call は描画されない。
  // 下のハイカード分岐は `owed === 0` を見ているのに、ここだけ抜けていた
  // (#4643 のレビュー指摘)。
  if (owed === 0) {
    return { targetAction: 'check', reason: 'frontendHint.followthequeenCheckFree', confidence: 'moderate' };
  }

  return hasHighCard(cards)
    ? { targetAction: 'call', reason: 'frontendHint.followthequeenCallHigh', confidence: 'moderate' }
    : { targetAction: 'fold', reason: 'frontendHint.followthequeenFoldWeak', confidence: 'moderate' };
}

/** クイーンのランク。常時ワイルド (sync: `FollowTheQueenQueenValue`)。 */
const QUEEN = 12;

/**
 * そのカードがいまワイルドか。クイーンは常にワイルドで、それ以外は
 * サーバが送る `wildRank` と一致したときだけ。
 *
 * `wildRank !== 0` は**防御的なガードで、テストでは観測できない** —— 52枚の
 * デッキに value 0 のカードは無いので、外しても答えは変わらない (ミューテーション
 * を当てて確認済み)。ドメイン側の `IsWild` も同じ形で、意図を揃えてある。
 */
function isWild(card: Card, wildRank: number): boolean {
  return card.value === QUEEN || (wildRank !== 0 && card.value === wildRank);
}

/** 同じランクが 2 枚以上あるか。 */
function hasPair(cards: Card[]): boolean {
  const seen = new Set<number>();
  for (const c of cards) {
    if (seen.has(c.value)) return true;
    seen.add(c.value);
  }
  return false;
}

/** 10 以上、または A を持っているか。 */
function hasHighCard(cards: Card[]): boolean {
  return cards.some((c) => c.value === ACE || c.value >= HIGH_CARD_FROM);
}
