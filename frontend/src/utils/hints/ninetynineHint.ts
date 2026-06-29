import type { Card, NinetyNineResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { NinetyNinePhase } from '../../types/phases';

/** Maps numeric trump suit to card design string. */
const SUIT_NUM_TO_DESIGN: Readonly<Record<number, Card['design']>> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** Returns a frontend HintResult for Ninety-Nine, or null if no suggestion. */
export function getNinetyNineHint(state: NinetyNineResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === NinetyNinePhase.BID) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return { targetAction: 'bury', reason: 'hint.buryStrategic', confidence: 'moderate' };
  }

  if (state.phase === NinetyNinePhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Hint for play phase: lead, follow suit, trump, or discard. */
function getPlayHint(cards: Card[], state: NinetyNineResponse): HintResult {
  const trumpDesign = SUIT_NUM_TO_DESIGN[state.trumpSuit];
  const trick = state.currentTrick;

  // Leading the trick
  if (trick.length === 0) {
    return { targetAction: 'play', reason: 'hint.leadStrategic', confidence: 'moderate' };
  }

  // Following suit
  const ledSuit = trick[0].card.design;
  const suitCards = cards.filter((c) => c.design === ledSuit);
  if (suitCards.length > 0) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }

  // Void in led suit: trump or discard
  const hasTrump = cards.some((c) => c.design === trumpDesign);
  if (hasTrump) {
    return { targetAction: 'play', reason: 'hint.trumpWithCard', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.discardLowest', confidence: 'moderate' };
}
