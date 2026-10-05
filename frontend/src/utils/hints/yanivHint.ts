import type { Card, YanivResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { YanivPhase } from '../../types/phases';
import { yanivCardValue } from '../yanivCombos';

/** Hand total at or below which Yaniv may be declared. */
const CALL_THRESHOLD = 5;

/** Sum of card values for a hand. */
function handTotal(cards: Card[]): number {
  return (cards ?? []).reduce((acc, c) => acc + yanivCardValue(c), 0);
}

/** Returns a Yaniv frontend hint or null. */
export function getYanivHint(state: YanivResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const human = state.players.find((p) => p.isHuman);
  if (!human || state.currentPlayerIdx !== human.id) return null;

  const cards = human.cards ?? [];

  if (state.phase === YanivPhase.DISCARD) {
    if (handTotal(cards) <= CALL_THRESHOLD) {
      return { targetAction: 'yaniv', reason: 'hint.declareYaniv', confidence: 'strong' };
    }
    return { targetAction: 'discard', reason: 'hint.discardHigh', confidence: 'moderate' };
  }

  if (state.phase === YanivPhase.DRAW) {
    const pickup = state.pickupCards ?? [];
    if (pickup.length > 0) {
      const ends = [pickup[0], pickup[pickup.length - 1]];
      const lowest = Math.min(...ends.map(yanivCardValue));
      if (lowest <= 2) {
        return { targetAction: 'drawpickup', reason: 'hint.takePickup', confidence: 'moderate' };
      }
    }
    return { targetAction: 'drawstock', reason: 'hint.drawStock', confidence: 'moderate' };
  }

  return null;
}
