import type { Card, SpadesResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SpadesPhase } from '../../types/phases';

/** High card threshold for bid estimation. */
const HIGH_CARD_VALUE = 11;
/** Spade suit design. */
const SPADE = 'SPADE';

/** Returns a frontend HintResult for Spades, or null if no suggestion. */
export function getSpadesHint(state: SpadesResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  if (state.phase === SpadesPhase.BID) {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getBidHint(human.cards);
  }

  if (state.phase === SpadesPhase.PLAY) {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Estimate bid count from high cards and spade count. */
function getBidHint(cards: Card[]): HintResult {
  const highCards = cards.filter((c) => c.value >= HIGH_CARD_VALUE).length;
  const spadeCount = cards.filter((c) => c.design === SPADE).length;
  const estimatedTricks = Math.max(1, Math.round(highCards * 0.7 + spadeCount * 0.3));
  const confidence = highCards >= 3 ? 'strong' : 'moderate';
  return { targetAction: `bid:${estimatedTricks}`, reason: 'hint.bidEstimate', confidence };
}

/** Hint for play phase: follow suit or play strategically. */
function getPlayHint(cards: Card[], state: SpadesResponse): HintResult {
  const trick = state.currentTrick;

  // Leading
  if (trick.length === 0) {
    if (!state.spadesBroken) {
      return { targetAction: 'play', reason: 'hint.leadNonSpade', confidence: 'strong' };
    }
    return { targetAction: 'play', reason: 'hint.leadStrategic', confidence: 'moderate' };
  }

  // Following
  const ledSuit = trick[0].card.design;
  const suitCards = cards.filter((c) => c.design === ledSuit);

  if (suitCards.length > 0) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }

  // Void in led suit: consider trumping with spade
  const hasSpades = cards.some((c) => c.design === SPADE);
  if (hasSpades) {
    return { targetAction: 'play', reason: 'hint.trumpWithSpade', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.discardLowest', confidence: 'moderate' };
}
