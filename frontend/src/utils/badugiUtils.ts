import type { Card } from '../types/card';

/**
 * Returns true when a 4-card Badugi hand is "complete" — all four cards have
 * different ranks AND different suits. This is the strongest possible Badugi
 * shape; exchanging any card can only make it weaker, so the UI surfaces a
 * stand-pat suggestion when this returns true.
 */
export function isCompleteBadugiHand(cards: ReadonlyArray<Card>): boolean {
  if (cards.length !== 4) return false;
  const ranks = new Set<number>();
  const suits = new Set<string>();
  for (const c of cards) {
    if (c.design === 'JOKER') return false;
    if (ranks.has(c.value) || suits.has(c.design)) return false;
    ranks.add(c.value);
    suits.add(c.design);
  }
  return true;
}
