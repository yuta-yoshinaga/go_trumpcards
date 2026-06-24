import type { Card } from '../types/card';

/**
 * Whether FreeCell's auto-complete is guaranteed to clear the entire board.
 *
 * The server's `AutoComplete` repeatedly sends any exposed (top-of-column or
 * free-cell) card to its foundation until nothing more can move. That process
 * fully clears the board exactly when every tableau column is **strictly
 * descending in rank from bottom to top**: then the globally smallest unplayed
 * card is always an exposed top card and is immediately playable, so the
 * collect can never stall. Free cells hold single exposed cards and never
 * block, so they don't affect the result.
 *
 * The check is intentionally conservative — equal ranks stacked across suits
 * are technically clearable but report `false`, so the badge never promises a
 * mop-up that could stall.
 *
 * @param tableau - The tableau columns (bottom card first; `null` slots ignored).
 * @returns `true` when auto-complete will deterministically win the game.
 */
export function freeCellAutoCompleteReady(tableau: readonly (Card | null)[][]): boolean {
  for (const col of tableau) {
    const cards = col.filter((c): c is Card => c !== null);
    for (let i = 0; i < cards.length - 1; i++) {
      // cards[i] sits below cards[i + 1]; the lower card must rank higher.
      if (cards[i].value <= cards[i + 1].value) return false;
    }
  }
  return true;
}
