import type { FreeCellMoveZone } from '../api/gameApi';
import type { Card } from '../types/card';
import { freeCellFoundationTarget } from './freeCellFoundationTarget';

/**
 * Computes the double-click auto-move target for `card` in Baker's Game.
 *
 * The FreeCell-family convenience move: first try the matching foundation pile
 * (same-suit ascending from Ace, via {@link freeCellFoundationTarget}); if no
 * legal foundation move exists, fall back to the first empty free cell (which
 * accepts any single card, mirroring the domain's `freeCells[cell] == nil`
 * check in `FreeCell.MoveTableauToFreeCell`). Returns `null` when neither a
 * foundation nor an empty free cell is available.
 *
 * @param card - The card being double-clicked (a tableau top card or a free-cell card).
 * @param foundation - The four foundation piles (bottom card first).
 * @param freeCells - The free-cell slots (`null` for empty).
 * @returns A `move` target zone, or `null` if no legal auto-move exists.
 */
export function bakersGameAutoMoveTarget(
  card: Card,
  foundation: Card[][],
  freeCells: (Card | null)[],
): FreeCellMoveZone | null {
  const foundationTarget = freeCellFoundationTarget(card, foundation);
  if (foundationTarget) return foundationTarget;

  const emptyCell = freeCells.findIndex((c) => c === null);
  return emptyCell === -1 ? null : { zone: 'freecell', cell: emptyCell };
}
