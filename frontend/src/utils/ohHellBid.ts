/** Summary of the table's placed bids relative to the round's hand size. */
export type OhHellBidSummary = {
  total: number;
  diff: number;
  kind: 'over' | 'under' | 'exact';
};

/**
 * Summarizes the placed bids for an Oh Hell round against the hand size.
 * In Oh Hell the dealer's hook forces the bid total to differ from the number
 * of tricks; this surfaces whether the table is currently over, under, or
 * exactly on the hand size.
 *
 * @param placedBids - Bids already declared this round (exclude undeclared players).
 * @param handSize - Number of cards dealt this round (tricks available).
 * @returns Total of placed bids, its difference from `handSize`, and the over/under/exact kind.
 */
export function ohHellBidSummary(placedBids: number[], handSize: number): OhHellBidSummary {
  const total = placedBids.reduce((sum, b) => sum + b, 0);
  const diff = total - handSize;
  return { total, diff, kind: diff > 0 ? 'over' : diff < 0 ? 'under' : 'exact' };
}
