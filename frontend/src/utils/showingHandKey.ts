import type { Card } from '../types/card';
import { evaluateBestHand, pokerHandKey } from './pokerSquaresUtils';

/** Return the hand translation key for visible cards, or null when none are visible. */
export function showingHandKey(doorCards: readonly Card[]): string | null {
  if (doorCards.length === 0) return null;
  const rank = evaluateBestHand(doorCards);
  return rank === null ? null : pokerHandKey(rank);
}
