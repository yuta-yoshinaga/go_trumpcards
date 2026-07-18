import type { FreeCellMoveZone } from '../api/gameApi';
import type { Card } from '../types/card';

/**
 * Maps a card design to its 0-based foundation index, matching the foundation
 * layout (`♠`=0, `♣`=1, `♥`=2, `♦`=3). Mirrors the backend's
 * `card.GetDesign() - 1` in `FreeCell.MoveTableauToFoundation`.
 */
const DESIGN_TO_FOUNDATION_INDEX: Record<string, number> = {
  SPADE: 0,
  CLOVER: 1,
  HEART: 2,
  DIAMOND: 3,
};

/**
 * Computes the legal foundation move target for `card` given the current
 * `foundation` piles, or `null` when no legal foundation move exists.
 *
 * Mirrors the domain's `canPlaceOnFoundation`: an empty pile accepts only an
 * Ace (`value === 1`); otherwise the card's suit must match the pile and its
 * rank must be exactly one higher than the pile's top card. The pile is
 * selected by suit, so a matching pile is always the same suit.
 *
 * @param card - The card being double-clicked (a tableau top card or a free-cell card).
 * @param foundation - The four foundation piles (bottom card first).
 * @returns A `{ zone: 'foundation', col }` move target, or `null` if illegal.
 */
export function freeCellFoundationTarget(card: Card, foundation: Card[][]): FreeCellMoveZone | null {
  const fIdx = DESIGN_TO_FOUNDATION_INDEX[card.design];
  if (fIdx === undefined || fIdx >= foundation.length) return null;
  const pile = foundation[fIdx];
  const placeable = pile.length === 0 ? card.value === 1 : card.value === pile[pile.length - 1].value + 1;
  return placeable ? { zone: 'foundation', col: fIdx } : null;
}
