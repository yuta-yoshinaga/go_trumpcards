import type { Card } from '../types/card';

/**
 * The suit and rank of Slobberhannes' penalty queen, mirroring
 * `SlobberhannesQueenSuit` / `SlobberhannesQueenValue` in
 * `internal/domain/Slobberhannes.go`.
 *
 * `TestSlobberhannesQueenMatchesTheDomain` compares these two literals against
 * the Go constants, so a change on either side fails rather than drifting.
 */
export const SLOBBERHANNES_QUEEN_DESIGN: Card['design'] = 'CLOVER';
/** Rank of the penalty queen. */
export const SLOBBERHANNES_QUEEN_VALUE = 12;

/**
 * Whether a card is the queen that costs a point to take.
 *
 * The other two penalties (first and last trick) depend on the trick *number*,
 * so the page can warn about them from `trickNumber` alone. This one depends on
 * what is on the table, which is why it needs a card test (#5745).
 * @param card - The card to check.
 * @returns True when it is the penalty queen.
 */
export function isPenaltyQueen(card: Card | null | undefined): boolean {
  return !!card && card.design === SLOBBERHANNES_QUEEN_DESIGN && card.value === SLOBBERHANNES_QUEEN_VALUE;
}
