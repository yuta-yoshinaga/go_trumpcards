import type { PenguinMoveZone } from '../api/gameApi';
import type { Card } from '../types/card';

/**
 * Maps a card design to its 0-based foundation index, matching Penguin's
 * `card.GetDesign() - 1` foundation selection (`♠`=0, `♣`=1, `♥`=2, `♦`=3).
 */
const DESIGN_TO_FOUNDATION_INDEX: Record<string, number> = {
  SPADE: 0,
  CLOVER: 1,
  HEART: 2,
  DIAMOND: 3,
};

/**
 * Computes the legal foundation move target for `card`, or `null` when no
 * legal foundation move exists. Unlike EightOff, an empty Penguin pile starts
 * at the deal's `baseRank`, and a non-empty pile wraps from King to Ace.
 *
 * @param card - The exposed card being double-clicked.
 * @param foundation - The four foundation piles (bottom card first).
 * @param baseRank - The rank that starts every empty foundation pile.
 * @returns A `{ zone: 'foundation', col }` move target, or `null` if illegal.
 */
export function penguinFoundationTarget(card: Card, foundation: Card[][], baseRank: number): PenguinMoveZone | null {
  const fIdx = DESIGN_TO_FOUNDATION_INDEX[card.design];
  if (fIdx === undefined || fIdx >= foundation.length) return null;
  const pile = foundation[fIdx];
  const placeable =
    pile.length === 0
      ? card.value === baseRank
      : card.design === pile[pile.length - 1].design && card.value === (pile[pile.length - 1].value % 13) + 1;
  return placeable ? { zone: 'foundation', col: fIdx } : null;
}
