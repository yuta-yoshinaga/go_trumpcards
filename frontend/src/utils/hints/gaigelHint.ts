import type { Card, GaigelResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { GaigelPhase } from '../../types/phases';

/** Maps numeric trump suit to card design string. */
const SUIT_NUM_TO_DESIGN: Readonly<Record<number, Card['design']>> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** Gaigel card-point values used to gauge a card's worth (A=11 … 7=0). */
const CARD_POINTS: Readonly<Record<number, number>> = {
  1: 11, // A
  10: 10,
  13: 4, // K
  12: 3, // Q
  11: 2, // J
  7: 0,
};

/** Returns the Gaigel card-point value for a card (0 when unknown). */
function cardPoints(card: Card): number {
  return CARD_POINTS[card.value] ?? 0;
}

/**
 * Returns a frontend HintResult for Gaigel, or null when no suggestion applies
 * (no human, empty hand, not the human's turn, or a non-play phase).
 */
export function getGaigelHint(state: GaigelResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (state.phase !== GaigelPhase.PLAY) return null;
  if (state.currentPlayerIdx !== humanIdx) return null;

  // Marriage takes priority on the lead when available.
  if (state.currentTrick.length === 0 && state.marriageIndices.length > 0) {
    return { targetAction: 'marriage', reason: 'hint.marriage', confidence: 'strong' };
  }

  return getPlayHint(human.cards, state);
}

/** Hint for play phase: lead, cut, win, or dump. */
function getPlayHint(cards: Card[], state: GaigelResponse): HintResult {
  const trumpDesign = SUIT_NUM_TO_DESIGN[state.trumpSuit];
  const trick = state.currentTrick;

  if (trick.length === 0) {
    const hasTrump = cards.some((c) => c.design === trumpDesign);
    if (hasTrump) {
      return { targetAction: 'play', reason: 'hint.leadTrump', confidence: 'moderate' };
    }
    const hasHighValue = cards.some((c) => cardPoints(c) >= 10);
    if (hasHighValue) {
      return { targetAction: 'play', reason: 'hint.leadValue', confidence: 'moderate' };
    }
    return { targetAction: 'play', reason: 'hint.leadLow', confidence: 'moderate' };
  }

  const ledSuit = trick[0].card.design;
  const hasLed = cards.some((c) => c.design === ledSuit);
  if (hasLed) {
    return { targetAction: 'play', reason: 'hint.followWin', confidence: 'strong' };
  }

  const hasTrump = cards.some((c) => c.design === trumpDesign);
  if (hasTrump) {
    return { targetAction: 'play', reason: 'hint.followCut', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.followDump', confidence: 'moderate' };
}
