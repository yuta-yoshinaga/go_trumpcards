import type { Card, FiveCardStudResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { FiveCardStudPhase } from '../../types/phases';

/** ベットのあるストリート。ここ以外では賭けの助言をしない。 */
const BETTING_STREETS: readonly number[] = [
  FiveCardStudPhase.SECOND_STREET,
  FiveCardStudPhase.THIRD_STREET,
  FiveCardStudPhase.FOURTH_STREET,
  FiveCardStudPhase.FIFTH_STREET,
];

/** 10 以上を「高い札」とみなす (10, J, Q, K, A)。A は 1 で届く。 */
const HIGH_CARD_FROM = 10;
const ACE = 1;

/**
 * Returns a frontend {@link HintResult} for Five Card Stud, or null when no
 * suggestion is available.
 *
 * **`handRank` is not used.** The presenter fills it only at showdown
 * (`FiveCardStudWebPresenter.buildPlayersOutput` gates it on `isShowdown`), and
 * this hint only runs on a betting street, so reading it would have made the
 * made-hand branch dead in production while the unit tests — which set the
 * field directly — went on passing. Found in review on #4622.
 *
 * So the hand is evaluated here from the cards the human can see, the way
 * `badugiHint` does. It is a pair check plus a high-card check: enough to
 * separate "worth paying for" from "not", without pretending to more.
 *
 * It also does **not** consult `maxBetAmount`. That is only positive under
 * Pot-Limit, and Five Card Stud is configured Fixed-Limit
 * (`FiveCardStudConfig.go:54`), so testing it would have gated the whole
 * raise branch on a value that is always zero.
 */
export function getFiveCardStudHint(state: FiveCardStudResponse): HintResult | null {
  if (state.gameEndFlag || !BETTING_STREETS.includes(state.phase)) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || human.folded || human.allIn || state.currentTurn !== human.id) return null;

  const cards = [...human.holeCards, ...human.doorCards];
  if (cards.length === 0) return null;

  // **既に払い込んでいる分は差し引く。**同額まで出していれば負債はない。
  const owed = Math.max(0, state.lastBet - human.currentBet);

  if (hasPair(cards)) {
    return { targetAction: 'raise', reason: 'frontendHint.fivecardstudRaisePair', confidence: 'moderate' };
  }

  if (owed === 0) {
    return { targetAction: 'check', reason: 'frontendHint.fivecardstudCheckFree', confidence: 'moderate' };
  }

  return hasHighCard(cards)
    ? { targetAction: 'call', reason: 'frontendHint.fivecardstudCallHigh', confidence: 'moderate' }
    : { targetAction: 'fold', reason: 'frontendHint.fivecardstudFoldWeak', confidence: 'moderate' };
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
