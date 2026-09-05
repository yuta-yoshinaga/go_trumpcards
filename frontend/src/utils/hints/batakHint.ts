import type { BatakResponse, Card } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BatakPhase } from '../../types/phases';

/** High card threshold for bid estimation. */
const HIGH_CARD_VALUE = 11;
/** Spade suit design (permanent trump). */
const SPADE = 'SPADE';

/** Returns a frontend HintResult for Batak, or null if no suggestion. */
export function getBatakHint(state: BatakResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === BatakPhase.BID) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getBidHint(human.cards, state.minLegalBid);
  }

  if (state.phase === BatakPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/**
 * Estimate bid count from high cards and spade count.
 * Minimum bid in Batak is 5 (or higher if another player bid).
 * If estimated tricks are less than minLegalBid (or minLegalBid is 0, meaning only pass is legal),
 * recommends pass instead of a number 0.
 */
function getBidHint(cards: Card[], minLegalBid: number = 5): HintResult {
  const highCards = cards.filter((c) => c.value >= HIGH_CARD_VALUE).length;
  const spadeCount = cards.filter((c) => c.design === SPADE).length;
  const estimatedTricks = Math.round(highCards * 0.7 + spadeCount * 0.3);
  const confidence = highCards >= 3 ? 'strong' : 'moderate';

  if (minLegalBid === 0 || estimatedTricks < minLegalBid) {
    return { targetAction: 'pass', reason: 'hint.passEstimate', confidence };
  }
  const bid = Math.min(13, estimatedTricks);
  return { targetAction: `bid:${bid}`, reason: 'hint.bidEstimate', confidence };
}

/**
 * Hint for play phase:
 *  - Lead → discourage spades if not broken, else play strongest.
 *  - Following with led suit available → follow.
 *  - Void → MUST trump if you have spades (Batak rule), otherwise discard.
 */
function getPlayHint(cards: Card[], state: BatakResponse): HintResult {
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

  // Void in led suit: Batak requires trumping if you have a spade.
  const hasSpades = cards.some((c) => c.design === SPADE);
  if (hasSpades) {
    return { targetAction: 'play', reason: 'hint.mustTrumpWithSpade', confidence: 'strong' };
  }

  return { targetAction: 'play', reason: 'hint.discardLowest', confidence: 'moderate' };
}
