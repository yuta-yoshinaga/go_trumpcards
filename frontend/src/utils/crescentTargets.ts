import type { Card, CrescentTableauCard } from '../types/card';

/** Highest rank in a standard deck (King). */
const CRESCENT_MAX_VALUE = 13;
/** Foundation indices 0..3 build up (A→K); 4..7 build down (K→A). */
const CRESCENT_ASCENDING_FOUNDATION_CNT = 4;

/**
 * Maps a card design to its 0-based suit index, matching the foundation layout
 * (`♠`=0, `♣`=1, `♥`=2, `♦`=3). Mirrors the backend's `crescentFoundationSuit`.
 */
const DESIGN_TO_SUIT_INDEX: Record<string, number> = {
  SPADE: 0,
  CLOVER: 1,
  HEART: 2,
  DIAMOND: 3,
};

/**
 * Reports whether `card` can be placed on foundation pile `fIdx`.
 * Ascending piles (0..3) build up A→K in-suit; descending piles (4..7) build
 * down K→A in-suit. Mirrors the domain's `canPlaceOnFoundation`.
 */
export function crescentCanPlaceOnFoundation(card: Card, foundation: Card[][], fIdx: number): boolean {
  if (fIdx < 0 || fIdx >= foundation.length) return false;
  if (DESIGN_TO_SUIT_INDEX[card.design] !== fIdx % CRESCENT_ASCENDING_FOUNDATION_CNT) return false;
  const pile = foundation[fIdx];
  const ascending = fIdx < CRESCENT_ASCENDING_FOUNDATION_CNT;
  if (pile.length === 0) {
    return ascending ? card.value === 1 : card.value === CRESCENT_MAX_VALUE;
  }
  const top = pile[pile.length - 1];
  return ascending ? card.value === top.value + 1 : card.value === top.value - 1;
}

/**
 * Reports whether `card` can be placed on the top of tableau column `toCol`:
 * same suit, adjacent rank (±1) with A↔K wrap allowed. Empty columns reject.
 * Mirrors the domain's `canPlaceOnTableau`.
 */
export function crescentCanPlaceOnTableau(card: Card, tableau: CrescentTableauCard[][], toCol: number): boolean {
  if (toCol < 0 || toCol >= tableau.length) return false;
  const col = tableau[toCol];
  if (col.length === 0) return false;
  const top = col[col.length - 1].card;
  if (!top || card.design !== top.design) return false;
  const cv = card.value;
  const tv = top.value;
  if (cv === tv + 1 || cv === tv - 1) return true;
  if (cv === 1 && tv === CRESCENT_MAX_VALUE) return true;
  if (cv === CRESCENT_MAX_VALUE && tv === 1) return true;
  return false;
}
