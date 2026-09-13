import type { Card } from '../types/card';

/** Suit designs in the numeric order used by CatchTen. */
const DESIGN_ORDER: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

/**
 * Returns the honor points for a Catch the Ten card.
 *
 * Sync: CatchTen.catchTenHonorPoints
 *
 * @param card - The card to score.
 * @param trumpSuit - The numeric trump suit, or 0 while undecided.
 * @returns The card's honor points.
 */
export function catchTenHonorPoints(card: Card | null | undefined, trumpSuit: number): number {
  if (!card || DESIGN_ORDER[card.design] !== trumpSuit) return 0;
  return { 11: 11, 10: 10, 1: 4, 13: 3, 12: 2 }[card.value] ?? 0;
}
