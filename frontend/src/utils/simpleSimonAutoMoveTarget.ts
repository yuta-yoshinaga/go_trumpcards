import type { Card } from '../types/card';

/**
 * Computes the double-click auto-move destination column for the run whose head
 * is `columns[fromCol][cardIndex]` in Simple Simon.
 *
 * Destinations are ranked by the priority from the game rules (mirroring the Go
 * domain's `SimpleSimon.canPlace`, which allows a card onto any column whose top
 * card is exactly one greater in value, suit ignored):
 *
 * 1. **Same-suit link** — a column whose top card is the same suit and one rank
 *    higher (builds toward a completable K→A run). Highest priority.
 * 2. **Rank-only link** — a column whose top card is one rank higher, any suit.
 * 3. **Empty column** — accepts any card, but only used as a fallback and only
 *    when the move does not merely relocate the whole source column (i.e. the
 *    run has cards above it, `cardIndex > 0`). Shuffling an entire pile onto an
 *    empty column exposes nothing, matching the domain's "not progress" rule in
 *    `hasAnyLegalMove`/`GetHint`.
 *
 * Within a tier the lowest column index wins. The source column is never a
 * destination.
 *
 * @param columns - The 10 tableau columns (top card last).
 * @param fromCol - Index of the source column.
 * @param cardIndex - Index of the run head within the source column.
 * @returns The destination column index, or `null` if no legal auto-move exists.
 */
export function simpleSimonAutoMoveTarget(columns: Card[][], fromCol: number, cardIndex: number): number | null {
  const head = columns[fromCol]?.[cardIndex];
  if (!head) return null;

  let rankOnly = -1;
  let empty = -1;
  for (let col = 0; col < columns.length; col++) {
    if (col === fromCol) continue;
    const pile = columns[col];
    if (pile.length === 0) {
      // Only offer an empty column when the move actually rearranges the board;
      // dropping an entire column onto an empty one makes no progress.
      if (empty === -1 && cardIndex > 0) empty = col;
      continue;
    }
    const top = pile[pile.length - 1];
    if (top.value === head.value + 1) {
      if (top.design === head.design) return col; // same-suit link: best target
      if (rankOnly === -1) rankOnly = col;
    }
  }
  if (rankOnly !== -1) return rankOnly;
  return empty === -1 ? null : empty;
}
