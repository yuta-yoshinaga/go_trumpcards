import type { AgnesTableauCard, Card } from '../types/card';

/**
 * Maps a card suit to its foundation pile index, mirroring the Go domain's
 * `fIdx = GetDesign() - 1` (Spade=1, Clover=2, Heart=3, Diamond=4).
 */
const DESIGN_TO_FOUNDATION: Record<string, number> = {
  SPADE: 0,
  CLOVER: 1,
  HEART: 2,
  DIAMOND: 3,
};

/** Next foundation rank, wrapping King (13) back to Ace (1). Mirrors domain `nextRank`. */
function nextRank(r: number): number {
  return (r % 13) + 1;
}

/** Face-up end (bottom) card of a tableau column, or null when empty/face-down. */
function endFaceUpCard(col: readonly AgnesTableauCard[]): Card | null {
  if (col.length === 0) return null;
  const tc = col[col.length - 1];
  return tc.faceUp ? tc.card : null;
}

/**
 * Whether a card can move to its suit's foundation pile: an empty pile takes the
 * base rank, otherwise it builds up in the same suit (with King→Ace wrap).
 * Mirrors the domain `canPlaceOnFoundation`.
 */
export function agnesCanPlaceOnFoundation(card: Card, foundation: readonly Card[][], baseRank: number): boolean {
  const idx = DESIGN_TO_FOUNDATION[card.design];
  if (idx === undefined) return false;
  const pile = foundation[idx] ?? [];
  if (pile.length === 0) return card.value === baseRank;
  const top = pile[pile.length - 1];
  return card.design === top.design && card.value === nextRank(top.value);
}

/**
 * Index of the first tableau column whose face-up end card can move to a
 * foundation, or -1 if none. Drives the auto-complete sweep by re-reading the
 * board after each move (mirrors the domain's foundation-first hint priority).
 */
export function agnesNextFoundationMove(
  tableau: readonly (readonly AgnesTableauCard[])[],
  foundation: readonly Card[][],
  baseRank: number,
): number {
  for (let col = 0; col < tableau.length; col++) {
    const card = endFaceUpCard(tableau[col]);
    if (card && agnesCanPlaceOnFoundation(card, foundation, baseRank)) return col;
  }
  return -1;
}
