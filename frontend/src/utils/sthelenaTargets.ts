import type { Card, StHelenaTableauCard } from '../types/card';

/** Highest rank in a standard deck (King). */
const STHELENA_MAX_VALUE = 13;
/** Foundation indices 0..3 build up (A→K); 4..7 build down (K→A). */
const STHELENA_ASCENDING_FOUNDATION_CNT = 4;
/** Columns 0..3 sit along the top of the circle, beside the king foundations. */
const STHELENA_TOP_COLUMN_CNT = 4;
/** Columns that touch both foundation rows, so they reach either one. */
const STHELENA_SIDE_COLUMNS = [4, 5, 10, 11];

/**
 * Maps a card design to its 0-based suit index, matching the foundation layout
 * (`♠`=0, `♣`=1, `♥`=2, `♦`=3). Mirrors the backend's `stHelenaFoundationSuit`.
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
export function stHelenaCanPlaceOnFoundation(card: Card, foundation: Card[][], fIdx: number): boolean {
  if (fIdx < 0 || fIdx >= foundation.length) return false;
  if (DESIGN_TO_SUIT_INDEX[card.design] !== fIdx % STHELENA_ASCENDING_FOUNDATION_CNT) return false;
  const pile = foundation[fIdx];
  const ascending = fIdx < STHELENA_ASCENDING_FOUNDATION_CNT;
  if (pile.length === 0) {
    return ascending ? card.value === 1 : card.value === STHELENA_MAX_VALUE;
  }
  const top = pile[pile.length - 1];
  return ascending ? card.value === top.value + 1 : card.value === top.value - 1;
}

/**
 * Reports whether column `fromCol` may send a card to foundation `fIdx` at all.
 *
 * On the first deal each column only reaches the foundation row it sits beside:
 * the top four reach the king row (4..7), the bottom four the ace row (0..3),
 * and the four side columns reach either. The first redeal lifts this, which is
 * what `restrictionsActive` reports. Mirrors the domain's `columnCanReach`.
 *
 * This is separate from {@link stHelenaCanPlaceOnFoundation}: a card can be the
 * right rank for a foundation its column cannot reach.
 */
export function stHelenaColumnCanReach(fromCol: number, fIdx: number, restrictionsActive: boolean): boolean {
  if (!restrictionsActive) return true;
  if (STHELENA_SIDE_COLUMNS.includes(fromCol)) return true;
  const ascending = fIdx < STHELENA_ASCENDING_FOUNDATION_CNT;
  return fromCol < STHELENA_TOP_COLUMN_CNT ? !ascending : ascending;
}

/**
 * Reports whether `card` can be placed on the top of tableau column `toCol`.
 *
 * **Rank only, either direction, and no king-ace wrap.** The clone source
 * (Crescent) builds in suit and wraps A↔K; keeping either here would hide legal
 * moves and offer illegal ones. Empty columns reject — they are never refilled.
 * Mirrors the domain's `canPlaceOnTableau`.
 */
export function stHelenaCanPlaceOnTableau(card: Card, tableau: StHelenaTableauCard[][], toCol: number): boolean {
  if (toCol < 0 || toCol >= tableau.length) return false;
  const col = tableau[toCol];
  if (col.length === 0) return false;
  const top = col[col.length - 1].card;
  if (!top) return false;
  return card.value === top.value + 1 || card.value === top.value - 1;
}
