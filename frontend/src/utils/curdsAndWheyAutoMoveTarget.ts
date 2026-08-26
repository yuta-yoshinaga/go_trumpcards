import type { Card } from '../types/card';

/**
 * Computes the double-click auto-move destination column for the run whose head
 * is `columns[fromCol][cardIndex]`.
 *
 * Mirrors the Go domain's `CurdsAndWhey.canPlace`. **This file duplicates the
 * server rule** -- change it whenever the domain changes. Simple Simon, which
 * this was cloned from, accepted any column whose top card was one rank higher
 * regardless of suit; Curds and Whey does not, so a "rank-only" tier here would
 * offer destinations the backend rejects.
 *
 * Destinations, in priority order:
 *
 * 1. **Same-suit link** -- top card is the same suit and one rank higher. This
 *    builds toward a completable K->A run, so it wins outright.
 * 2. **Same-rank link** -- top card has the same rank, any suit. Legal, but it
 *    only parks a card, so it ranks below a suit link.
 * 3. **Empty column** -- accepts any card, used only as a fallback and only when
 *    the move actually rearranges the board (`cardIndex > 0`). Dropping a whole
 *    column onto an empty one exposes nothing, matching the domain's
 *    "not progress" rule in `hasAnyLegalMove`/`GetHint`.
 *
 * Within a tier the lowest column index wins. The source column is never a
 * destination.
 *
 * @param columns - The 13 tableau columns (top card last).
 * @param fromCol - Index of the source column.
 * @param cardIndex - Index of the run head within the source column.
 * @returns The destination column index, or `null` if no legal auto-move exists.
 */
export function curdsAndWheyAutoMoveTarget(columns: Card[][], fromCol: number, cardIndex: number): number | null {
  const head = columns[fromCol]?.[cardIndex];
  if (!head) return null;

  let sameRank = -1;
  let empty = -1;
  for (let col = 0; col < columns.length; col++) {
    if (col === fromCol) continue;
    const pile = columns[col];
    if (pile.length === 0) {
      // Only offer an empty column when the move actually rearranges the board.
      if (empty === -1 && cardIndex > 0) empty = col;
      continue;
    }
    const top = pile[pile.length - 1];
    if (top.design === head.design && top.value === head.value + 1) return col;
    if (top.value === head.value && sameRank === -1) sameRank = col;
  }
  if (sameRank !== -1) return sameRank;
  return empty === -1 ? null : empty;
}
