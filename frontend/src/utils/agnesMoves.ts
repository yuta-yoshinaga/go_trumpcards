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

/** Whether a card is a black suit (Spade or Clover), matching the domain `isBlack`. */
function isBlack(card: Card): boolean {
  return card.design === 'SPADE' || card.design === 'CLOVER';
}

/** Whether two cards share a color, matching the domain `isSameColor`. */
function sameColor(a: Card, b: Card): boolean {
  return isBlack(a) === isBlack(b);
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
 * Whether a card can stack on a tableau column's end card: same color, one rank
 * lower, no wrap. Empty columns cannot be filled manually. Mirrors the domain
 * `canPlaceOnTableau`.
 */
export function agnesCanPlaceOnTableau(
  card: Card,
  tableau: readonly (readonly AgnesTableauCard[])[],
  toCol: number,
): boolean {
  const col = tableau[toCol];
  if (!col || col.length === 0) return false;
  const top = col[col.length - 1]?.card;
  if (!top) return false;
  return sameColor(card, top) && card.value === top.value - 1;
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

/**
 * Whether any legal move remains: a pending deal, a tableau→foundation move, or a
 * tableau→tableau move. Used to detect a stalemate so the UI can prompt the
 * player to undo or give up. Mirrors the union of the domain's `DealStock`
 * availability and `GetHint` move search.
 */
export function agnesHasLegalMove(
  tableau: readonly (readonly AgnesTableauCard[])[],
  foundation: readonly Card[][],
  baseRank: number,
  stockCount: number,
): boolean {
  if (stockCount > 0) return true;
  for (let col = 0; col < tableau.length; col++) {
    const card = endFaceUpCard(tableau[col]);
    if (!card) continue;
    if (agnesCanPlaceOnFoundation(card, foundation, baseRank)) return true;
    for (let to = 0; to < tableau.length; to++) {
      if (to !== col && agnesCanPlaceOnTableau(card, tableau, to)) return true;
    }
  }
  return false;
}
