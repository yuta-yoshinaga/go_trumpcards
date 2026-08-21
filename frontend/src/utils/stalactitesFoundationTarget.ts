import type { StalactitesMoveZone } from '../api/gameApi';
import type { Card } from '../types/card';

/**
 * Computes the legal foundation move target for `card`, or `null` when no legal
 * foundation move exists.
 *
 * Mirrors the domain's `Stalactites.canPlaceOnFoundation` and
 * `foundationIndexFor`. **This file duplicates the server rule** -- change it
 * whenever the domain rule changes. The FreeCell version this was cloned from
 * picked the pile by suit and required an Ace on an empty pile, which is wrong
 * for Stalactites in three ways:
 *
 *   - suit is IGNORED, so the pile cannot be chosen by design
 *   - an empty pile takes the deal's `baseRank`, not an Ace
 *   - the sequence WRAPS: Ace follows King
 *
 * Continuing an existing pile is preferred over opening an empty one, matching
 * the domain, so a card never opens a new pile when it could extend one.
 *
 * @param card - The card being double-clicked (a tableau top card or a cell card).
 * @param foundation - The four foundation piles (bottom card first).
 * @param baseRank - The rank every foundation starts from, from the API response.
 * @returns A `{ zone: 'foundation', col }` move target, or `null` if illegal.
 */
export function stalactitesFoundationTarget(
  card: Card,
  foundation: Card[][],
  baseRank: number,
): StalactitesMoveZone | null {
  const nextRank = (v: number): number => (v >= 13 ? 1 : v + 1);
  const accepts = (pile: Card[]): boolean =>
    pile.length === 0 ? card.value === baseRank : card.value === nextRank(pile[pile.length - 1].value);

  for (let i = 0; i < foundation.length; i++) {
    if (foundation[i].length > 0 && accepts(foundation[i])) return { zone: 'foundation', col: i };
  }
  for (let i = 0; i < foundation.length; i++) {
    if (foundation[i].length === 0 && accepts(foundation[i])) return { zone: 'foundation', col: i };
  }
  return null;
}
