/** The direction in which a Bisley foundation is built. */
export type BisleyFoundationDirection = 'ascending' | 'descending';

const RANKS_PER_SUIT = 13;

/**
 * Return the next rank required by a Bisley foundation, or null when complete.
 * @param topValue The rank at the top of the foundation in this direction.
 * @param suitCardsPlaced The total in the ascending and descending foundations; Bisley splits one suit across two foundations.
 * @param direction The direction in which this foundation is built.
 */
export function bisleyNextRank(
  topValue: number | undefined,
  suitCardsPlaced: number,
  direction: BisleyFoundationDirection,
): number | null {
  if (suitCardsPlaced >= RANKS_PER_SUIT) return null;
  if (topValue === undefined) return direction === 'ascending' ? 1 : 13;
  return direction === 'ascending' ? topValue + 1 : topValue - 1;
}
