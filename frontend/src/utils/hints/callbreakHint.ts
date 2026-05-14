import type { CallBreakResponse, Card } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CallBreakPhase } from '../../types/phases';

/** High card threshold for bid estimation. */
const HIGH_CARD_VALUE = 11;
/** Spade suit design (permanent trump). */
const SPADE = 'SPADE';

/** Returns a frontend HintResult for Call Break, or null if no suggestion. */
export function getCallBreakHint(state: CallBreakResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === CallBreakPhase.BID) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getBidHint(human.cards);
  }

  if (state.phase === CallBreakPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Estimate bid count from high cards and spade count. Minimum 1 (no Nil bid). */
function getBidHint(cards: Card[]): HintResult {
  const highCards = cards.filter((c) => c.value >= HIGH_CARD_VALUE).length;
  const spadeCount = cards.filter((c) => c.design === SPADE).length;
  const estimatedTricks = Math.max(1, Math.round(highCards * 0.7 + spadeCount * 0.3));
  const confidence = highCards >= 3 ? 'strong' : 'moderate';
  return { targetAction: `bid:${estimatedTricks}`, reason: 'hint.bidEstimate', confidence };
}

/**
 * Hint for play phase:
 *  - Lead → discourage spades if not broken, else play strongest.
 *  - Following with led suit available → follow.
 *  - Void → MUST trump if you have spades (Call Break rule), otherwise discard.
 */
function getPlayHint(cards: Card[], state: CallBreakResponse): HintResult {
  const trick = state.currentTrick;

  if (trick.length === 0) {
    if (!state.spadesBroken) {
      return { targetAction: 'play', reason: 'hint.leadNonSpade', confidence: 'strong' };
    }
    return { targetAction: 'play', reason: 'hint.leadStrategic', confidence: 'moderate' };
  }

  const ledSuit = trick[0].card.design;
  const suitCards = cards.filter((c) => c.design === ledSuit);

  if (suitCards.length > 0) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }

  // Void in led suit: Call Break requires trumping if you have a spade.
  const hasSpades = cards.some((c) => c.design === SPADE);
  if (hasSpades) {
    return { targetAction: 'play', reason: 'hint.mustTrumpWithSpade', confidence: 'strong' };
  }

  return { targetAction: 'play', reason: 'hint.discardLowest', confidence: 'moderate' };
}
