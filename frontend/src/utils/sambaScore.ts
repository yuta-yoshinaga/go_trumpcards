import type { Card } from '../types/card';

/**
 * Minimum initial-meld value required at the team's cumulative score,
 * mirroring the backend `Samba.minimumMeldValue` bands.
 *
 * @param cumulativeScore - The player's team cumulative score.
 * @returns The minimum points the first meld must total (15/50/90/120).
 */
export function sambaMinMeld(cumulativeScore: number): number {
  if (cumulativeScore < 0) return 15;
  if (cumulativeScore < 1500) return 50;
  if (cumulativeScore < 3000) return 90;
  return 120;
}

/**
 * Point value of a single card, mirroring the backend `SambaCardValue`.
 *
 * @param card - The card to score.
 * @returns Its Samba point value.
 */
export function sambaCardValue(card: Card): number {
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
 * @returns The summed Samba point value.
 */
export function sambaSelectionPoints(cards: Card[]): number {
  return cards.reduce((sum, card) => sum + sambaCardValue(card), 0);
}
