/** Minimal player shape needed to work out who a swap would hit. */
export interface CuckooSwapPlayer {
  id: number;
  isEliminated: boolean;
}

/**
 * Seat a swap from `fromIdx` would trade with: the next player round the table
 * who is still in. Returns `null` when nobody else is active, which is the case
 * the domain treats as keeping the card instead (`attemptSwap`).
 *
 * The dealer swaps with the stock rather than a neighbour, so callers handle
 * that case before asking. Mirrors `nextActiveIdx` in internal/domain/Cuckoo.go.
 *
 * @param players - Seats in table order.
 * @param fromIdx - Index of the player acting.
 * @returns The target seat index, or null when the swap would be a keep.
 */
export function cuckooSwapTarget(players: readonly CuckooSwapPlayer[], fromIdx: number): number | null {
  if (fromIdx < 0 || fromIdx >= players.length) return null;
  for (let step = 1; step < players.length; step++) {
    const idx = (fromIdx + step) % players.length;
    if (!players[idx].isEliminated) return idx;
  }
  return null;
}
