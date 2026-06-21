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
