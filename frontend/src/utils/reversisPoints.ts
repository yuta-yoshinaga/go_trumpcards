import type { Card } from '../types/card';

/**
 * Penalty a card costs the player who takes the trick it falls in.
 *
 * Mirrors `ReversisCardPenalty` in `internal/domain/Reversis.go`; the shared
 * golden vectors in `__fixtures__/reversisPoints.golden.json` are asserted from
 * both sides, so the badge cannot disagree with the score (#5747).
 */
export const REVERSIS_CARD_POINTS: Readonly<Record<number, number>> = {
  1: 4, // A
  13: 3, // K
  12: 2, // Q
  11: 1, // J
};

/**
 * Points a card is worth to whoever takes it, 0 for the plain ranks.
 * @param card - The card to score.
 * @returns 0-4.
 */
export function reversisCardPoints(card: Card | null | undefined): number {
  if (!card) return 0;
  return REVERSIS_CARD_POINTS[card.value] ?? 0;
}
