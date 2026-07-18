/** Progress of a Spades player toward their bid contract. */
export type SpadesBidProgress =
  | { kind: 'nilOk' }
  | { kind: 'nilFail' }
  | { kind: 'remaining'; remaining: number }
  | { kind: 'made'; bags: number };

/**
 * Derives a player's bid-contract progress from their bid and tricks won.
 * A Nil bid (0) is `nilOk` until the first trick, then `nilFail`. A positive
 * bid is `remaining` until met, then `made` with the overflow counted as bags.
 *
 * @param bid - The player's bid (0 means Nil).
 * @param trickCount - Tricks the player has won so far this round.
 * @returns The current bid progress descriptor.
 */
export function spadesBidProgress(bid: number, trickCount: number): SpadesBidProgress {
  if (bid <= 0) {
    return trickCount > 0 ? { kind: 'nilFail' } : { kind: 'nilOk' };
  }
  if (trickCount < bid) {
    return { kind: 'remaining', remaining: bid - trickCount };
  }
  return { kind: 'made', bags: trickCount - bid };
}

/** Number of bags within the penalty threshold that triggers a warning. */
export const SPADES_BAG_WARNING_MARGIN = 2;

/**
 * Severity of a Spades player's bag count as it nears the penalty threshold.
 * `warn` when the player is exactly `SPADES_BAG_WARNING_MARGIN` bags away;
 * `danger` once within one bag of (or having reached) the threshold.
 */
export type SpadesBagWarning = { level: 'warn' | 'danger'; bags: number; threshold: number };

/**
 * Determines whether a player's bag count is close enough to the penalty
 * threshold to warrant a warning. Returns `null` when the threshold is disabled
 * (`<= 0`) or the player is more than {@link SPADES_BAG_WARNING_MARGIN} bags
 * away. Within the margin the level escalates: `warn` at the margin edge,
 * `danger` at one bag away or once the threshold is reached.
 *
 * @param bags - The player's accumulated bags (overtricks).
 * @param threshold - The bag penalty threshold (`config.bagPenaltyThreshold`).
 * @returns A warning descriptor, or `null` when no warning applies.
 */
export function spadesBagWarning(bags: number, threshold: number): SpadesBagWarning | null {
  if (threshold <= 0 || bags < threshold - SPADES_BAG_WARNING_MARGIN) {
    return null;
  }
  const level = bags >= threshold - 1 ? 'danger' : 'warn';
  return { level, bags, threshold };
}
