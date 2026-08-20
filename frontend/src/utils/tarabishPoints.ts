import type { Card } from '../types/card';

/** Card values by rank when the card is NOT the trump suit. */
const PLAIN_POINTS: Readonly<Record<number, number>> = { 1: 11, 10: 10, 13: 4, 12: 3, 11: 2 };
/** Card values by rank when the card IS the trump suit — the Jass and the Menel jump. */
const TRUMP_POINTS: Readonly<Record<number, number>> = { 11: 20, 9: 14, 1: 11, 10: 10, 13: 4, 12: 3 };

/** Suit designs in the numeric order the backend uses (1-based, matching Card.GetDesign). */
const DESIGN_ORDER: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

/**
 * What a card is worth in the current hand.
 *
 * **Only the trump suit swaps tables** (J=20 Jass, 9=14 Menel), which is the
 * whole point of this family — the same jack is worth 20 or 2 depending on the
 * suit called, so a hand read without it invites feeding the wrong card to a
 * partner (#5749). Mirrors `TarabishCardPoints` in `internal/domain/Tarabish.go`;
 * the golden vectors in `__fixtures__/tarabishPoints.golden.json` are asserted
 * from both sides.
 * @param card - The card to score.
 * @param trumpSuit - The trump suit as a design constant (0 while undecided).
 * @returns The card's point value.
 */
export function tarabishCardPoints(card: Card | null | undefined, trumpSuit: number): number {
  if (!card) return 0;
  const design = DESIGN_ORDER[card.design] ?? -1;
  const table = design === trumpSuit ? TRUMP_POINTS : PLAIN_POINTS;
  return table[card.value] ?? 0;
}
