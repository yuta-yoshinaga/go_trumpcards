import type { Card, FiftyOneResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Score threshold above which we recommend calling stop. */
const STOP_THRESHOLD = 40;

/** Minimum improvement to recommend an exchange with strong confidence. */
const STRONG_IMPROVEMENT = 5;

/** Map Card.design to a suit key for grouping. */
const DESIGN_TO_SUIT: Record<Card['design'], string> = {
  SPADE: 'S',
  CLOVER: 'C',
  HEART: 'H',
  DIAMOND: 'D',
  JOKER: '',
};

/** Card face value for scoring (Ace = 11, J/Q/K = 10). */
function cardScore(value: number): number {
  if (value === 1) return 11;
  if (value >= 10) return 10;
  return value;
}

/** Calculate the best single-suit score from a set of cards. */
function bestSuitScore(cards: Card[]): number {
  const suitTotals: Record<string, number> = {};
  for (const c of cards) {
    const suit = DESIGN_TO_SUIT[c.design];
    if (!suit) continue;
    suitTotals[suit] = (suitTotals[suit] ?? 0) + cardScore(c.value);
  }
  let best = 0;
  for (const total of Object.values(suitTotals)) {
    if (total > best) best = total;
  }
  return best;
}

/** Returns a Fifty-one frontend hint or null. */
export function getFiftyOneHint(state: FiftyOneResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || state.currentTurn !== human.id) return null;

  const currentScore = bestSuitScore(human.cards);

  // If score is already high, recommend stopping.
  if (currentScore >= STOP_THRESHOLD && state.stopCallerIdx < 0) {
    return { targetAction: 'stop', reason: 'hint.stopHigh', confidence: 'strong' };
  }

  // Find the best single-card exchange.
  let bestImprovement = 0;
  for (let hi = 0; hi < human.cards.length; hi++) {
    for (let ti = 0; ti < state.tableCards.length; ti++) {
      const newHand = [...human.cards];
      newHand[hi] = state.tableCards[ti];
      const newScore = bestSuitScore(newHand);
      const improvement = newScore - currentScore;
      if (improvement > bestImprovement) {
        bestImprovement = improvement;
      }
    }
  }

  if (bestImprovement > 0) {
    return {
      targetAction: 'exchange',
      reason: bestImprovement >= STRONG_IMPROVEMENT ? 'hint.exchangeStrong' : 'hint.exchangeModerate',
      confidence: bestImprovement >= STRONG_IMPROVEMENT ? 'strong' : 'moderate',
    };
  }

  return null;
}
