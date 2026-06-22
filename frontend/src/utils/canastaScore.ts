import type { Card } from '../types/card';

/**
 * Minimum initial-meld value required at the player's cumulative score,
 * mirroring the backend `minimumMeldValue` bands.
 *
 * @param cumulativeScore - The player's cumulative score.
 * @returns The minimum points the first meld must total (15/50/90/120).
 */
export function canastaMinMeld(cumulativeScore: number): number {
  if (cumulativeScore < 0) return 15;
  if (cumulativeScore < 1500) return 50;
  if (cumulativeScore < 3000) return 90;
  return 120;
}

/**
 * Point value of a single card, mirroring the backend `CanastaCardValue`.
 *
 * @param card - The card to score.
 * @returns Its Canasta point value.
 */
export function canastaCardValue(card: Card): number {
  if (card.design === 'JOKER') return 50;
  if (card.value === 2 || card.value === 1) return 20;
  if (card.value === 3 && (card.design === 'SPADE' || card.design === 'CLOVER')) return 5;
  if (card.value >= 8) return 10;
  return 5;
}

/**
 * Total point value of a set of selected cards.
 *
 * @param cards - The selected cards.
 * @returns The summed Canasta point value.
 */
export function canastaSelectionPoints(cards: Card[]): number {
  return cards.reduce((sum, card) => sum + canastaCardValue(card), 0);
}
