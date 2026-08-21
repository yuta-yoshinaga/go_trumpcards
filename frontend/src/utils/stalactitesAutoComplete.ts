import type { Card } from '../types/card';

/**
 * Whether Stalactites's auto-complete is guaranteed to clear the entire board.
 *
 * The server's `AutoComplete` repeatedly sends any exposed (top-of-column or
 * cell) card to a foundation until nothing more can move. That process fully
 * clears the board exactly when every tableau column is **strictly descending
 * in foundation order from bottom to top**: then the next card the foundations
 * want is always an exposed top card, so the collect can never stall. Cells
 * hold single exposed cards and never block, so they don't affect the result.
 *
 * **Foundation order is not rank order.** Stalactites starts every pile at the
 * deal's `baseRank` and wraps King -> Ace, so the sequence for base 7 is
 * 7,8,...,K,A,2,...,6. Comparing raw ranks -- as the FreeCell version this was
 * cloned from did -- calls a column like [K, A] "descending" and promises a
 * mop-up, when in fact the King must be played first and is buried underneath.
 *
 * The check is intentionally conservative: equal ranks stacked across suits are
 * technically clearable but report `false`, so the badge never promises a
 * mop-up that could stall.
 *
 * @param tableau - The tableau columns (bottom card first; `null` slots ignored).
 * @param baseRank - The rank every foundation starts from, from the API response.
 * @returns `true` when auto-complete will deterministically win the game.
 */
export function stalactitesAutoCompleteReady(tableau: readonly (Card | null)[][], baseRank: number): boolean {
  /** How far a rank sits along the foundation sequence: 0 is played first. */
  const order = (c: Card): number => (c.value - baseRank + 13) % 13;

  for (const col of tableau) {
    const cards = col.filter((c): c is Card => c !== null);
    for (let i = 0; i < cards.length - 1; i++) {
      // cards[i] sits below cards[i + 1]; the lower card must be played later.
      if (order(cards[i]) <= order(cards[i + 1])) return false;
    }
  }
  return true;
}
