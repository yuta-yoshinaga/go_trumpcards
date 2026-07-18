import type { Card } from '../types/card';

/**
 * Computes the set of foundation pile indices that `card` can legally move to,
 * given the current `foundation` piles.
 *
 * Mirrors the domain's `FortyAndEight.canPlaceOnFoundation`: an empty pile
 * accepts only an Ace (`value === 1`); a non-empty pile accepts a card of the
 * same suit whose rank is exactly one higher than the pile's top card. Unlike
 * suit-fixed solitaires, Forty and Eight's eight foundations are not bound to a
 * suit, so an Ace is eligible for every empty pile.
 *
 * @param card - The selected source card (waste top card or a tableau card), or `null`.
 * @param foundation - The eight foundation piles (bottom card first).
 * @returns A set of 0-based foundation indices the card can be placed on (empty when none).
 */
export function fortyAndEightFoundationTargets(card: Card | null | undefined, foundation: Card[][]): Set<number> {
  const targets = new Set<number>();
  if (!card) return targets;
  foundation.forEach((pile, idx) => {
    const placeable =
      pile.length === 0
        ? card.value === 1
        : card.design === pile[pile.length - 1].design && card.value === pile[pile.length - 1].value + 1;
    if (placeable) targets.add(idx);
  });
  return targets;
}
