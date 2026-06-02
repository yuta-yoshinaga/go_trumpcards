import type { Card, ThirtyOneResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ThirtyOnePhase } from '../../types/phases';

/** Best-suit score at or above which we recommend knocking. */
const KNOCK_THRESHOLD = 27;

/** Card score in Thirty-One: Ace = 11, J/Q/K = 10, else face value. */
function cardScore(value: number): number {
  if (value === 1) return 11;
  if (value >= 11) return 10;
  return value;
}

/** Highest single-suit total for a set of cards. */
function bestSuitScore(cards: Card[]): number {
  if (!cards) return 0;
  const totals: Record<string, number> = {};
  for (const c of cards) {
    if (!c || c.design === 'JOKER') continue;
    totals[c.design] = (totals[c.design] ?? 0) + cardScore(c.value);
  }
  let best = 0;
  for (const total of Object.values(totals)) {
    if (total > best) best = total;
  }
  return best;
}

/** Returns a Thirty-One frontend hint or null. */
export function getThirtyOneHint(state: ThirtyOneResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || state.currentPlayerIdx !== human.id) return null;
  if (state.phase !== ThirtyOnePhase.DRAW) return null;

  const cards = human.cards ?? [];
  const currentScore = bestSuitScore(cards);

  // Strong hand and nobody has knocked yet — recommend knocking.
  if (currentScore >= KNOCK_THRESHOLD && state.knockerIdx < 0) {
    return { targetAction: 'knock', reason: 'hint.knockHigh', confidence: 'strong' };
  }

  // Does taking the discard top improve the best-suit score?
  if (state.discardTop) {
    let bestWithDiscard = currentScore;
    for (let i = 0; i < cards.length; i++) {
      const newHand = [...cards];
      newHand[i] = state.discardTop;
      const s = bestSuitScore(newHand);
      if (s > bestWithDiscard) bestWithDiscard = s;
    }
    if (bestWithDiscard > currentScore) {
      return { targetAction: 'drawdiscard', reason: 'hint.drawDiscard', confidence: 'moderate' };
    }
  }

  return { targetAction: 'drawstock', reason: 'hint.drawStock', confidence: 'moderate' };
}
