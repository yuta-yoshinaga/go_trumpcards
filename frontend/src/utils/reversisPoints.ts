import type { Card } from '../types/card';

/**
 * Base penalty by rank, mirroring `ReversisCardPenalty` in
 * `internal/domain/Reversis.go`.
 */
export const REVERSIS_CARD_POINTS: Readonly<Record<number, number>> = {
  1: 4, // A
  13: 3, // K
  12: 2, // Q
  11: 1, // J
};

/** Extra penalty charged for capturing a marked card (`ReversisMarkedPenalty`). */
export const REVERSIS_MARKED_PENALTY = 5;

/** True for the Quinola (heart jack) — one of the two marked cards. */
function isQuinola(card: Card): boolean {
  return card.design === 'HEART' && card.value === 11;
}

/** True for the diamond ace — the other marked card. */
function isDiamondAce(card: Card): boolean {
  return card.design === 'DIAMOND' && card.value === 1;
}

/**
 * What a card actually costs whoever takes the trick it falls in.
 *
 * **The two marked cards are not just their rank.** The Quinola and the diamond
 * ace add `REVERSIS_MARKED_PENALTY` on top (`chargeMarked` in the domain), so a
 * badge built from the rank alone would understate exactly the two cards this
 * label matters most for (#5747). The shared golden vectors in
 * `__fixtures__/reversisPoints.golden.json` are asserted from both sides.
 * @param card - The card to score.
 * @returns The penalty its captor pays.
 */
export function reversisCardPoints(card: Card | null | undefined): number {
  if (!card) return 0;
  const base = REVERSIS_CARD_POINTS[card.value] ?? 0;
  return isQuinola(card) || isDiamondAce(card) ? base + REVERSIS_MARKED_PENALTY : base;
}
