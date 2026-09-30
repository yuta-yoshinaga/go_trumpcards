import type { Card } from '../types/card';

/** Returns the shared Canasta-family point value of one card. */
export function canastaFamilyCardValue(card: Card): number {
  if (card.design === 'JOKER') return 50;
  if (card.value === 2 || card.value === 1) return 20;
  if (card.value === 3 && (card.design === 'SPADE' || card.design === 'CLOVER')) return 5;
  if (card.value >= 8) return 10;
  return 5;
}

/** Sums the shared Canasta-family point values of selected cards. */
export function canastaFamilySelectionPoints(cards: Card[]): number {
  return cards.reduce((sum, card) => sum + canastaFamilyCardValue(card), 0);
}
