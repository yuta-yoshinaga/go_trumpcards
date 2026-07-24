import type { Card } from '../types/card';
import { evaluatePartialHand, PokerHand } from './pokerSquaresUtils';

/**
 * Rough win-probability tier for a Guts declaration guide. This is a simple
 * heuristic label (not a Monte-Carlo estimate) meant to nudge the In/Out call.
 */
export type GutsGuideTier = 'high' | 'medium' | 'low';

/** A Guts declaration guide: the named hand plus a rough win-chance tier. */
export interface GutsHandGuide {
  /** i18n suffix for the hand name (`'pair'` or `'highcard'`). */
  handKey: 'pair' | 'highcard';
  /** Rough win-probability tier. */
  tier: GutsGuideTier;
}

/** Ace-high rank value: a low Ace (value 1) counts as 14, the strongest rank. */
function highRank(value: number): number {
  return value === 1 ? 14 : value;
}

/**
 * Evaluate the human's Guts hand into a declaration guide: names the hand
 * (pair vs high card) via the shared poker partial-hand evaluator and assigns a
 * rough win-probability tier. Any pair is a strong ("high") hand; an unpaired
 * hand is "medium" when its top card is a King or Ace, otherwise "low".
 *
 * @param cards - The human's hand (typically 2 cards in Guts).
 * @returns A {@link GutsHandGuide}, or `null` for an empty hand.
 */
export function evaluateGutsGuide(cards: readonly Card[]): GutsHandGuide | null {
  if (cards.length === 0) return null;

  const made = evaluatePartialHand(cards);
  const hasPair = made !== null && made >= PokerHand.OnePair;
  if (hasPair) {
    return { handKey: 'pair', tier: 'high' };
  }

  const topRank = Math.max(...cards.map((c) => highRank(c.value)));
  const tier: GutsGuideTier = topRank >= 13 ? 'medium' : 'low';
  return { handKey: 'highcard', tier };
}
