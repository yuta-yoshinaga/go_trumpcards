import type { FortyThievesMoveZone } from '../api/gameApi';
import type { Card } from '../types/card';

/**
 * Computes the legal foundation move target for `card` given the current
 * `foundation` piles, or `null` when no legal foundation move exists.
 *
 * Mirrors the domain's `findFoundation` + `canPlaceOnFoundation` in
 * `FortyThieves.go`: the eight foundation piles are NOT locked to a suit by
 * index, so they are scanned in order and the first pile that accepts the card
 * wins. An empty pile accepts only an Ace (`value === 1`); a non-empty pile
 * requires the same suit and a rank exactly one higher than its top card.
 *
 * The backend recomputes the destination pile itself (`MoveTableauToFoundation`
 * / `MoveWasteToFoundation` take no foundation index), so the returned `col`
 * only needs to identify a legal move — its exact value is not sent as the
 * destination. It is returned for symmetry with the tableau/foundation zone
 * shape and for hint-ring reuse.
 *
 * @param card - The exposed waste or tableau top card being double-clicked.
 * @param foundation - The eight foundation piles (bottom card first).
 * @returns A `{ zone: 'foundation', col }` move target, or `null` if illegal.
 */
export function fortyThievesFoundationTarget(card: Card, foundation: Card[][]): FortyThievesMoveZone | null {
  for (let i = 0; i < foundation.length; i++) {
    const pile = foundation[i];
    const top = pile.length > 0 ? pile[pile.length - 1] : null;
    const placeable = top === null ? card.value === 1 : card.design === top.design && card.value === top.value + 1;
    if (placeable) return { zone: 'foundation', col: i };
  }
  return null;
}
