import type { Card } from '../types/card';

/** Index sets identifying a player's qualifying low cards within their hole and the board. */
export interface LowCardIndexSets {
  loHoleSet: Set<number>;
  loBoardSet: Set<number>;
}

/** Match two cards by design and value (single deck → unique). */
function sameCard(a: Card, b: Card): boolean {
  return a.design === b.design && a.value === b.value;
}

/**
 * Maps an Omaha Hi-Lo qualifying low hand to the indices of the matching cards
 * in the player's hole and the community board, so they can be highlighted.
 *
 * @param lowBestHand - The five cards forming the qualifying low (or undefined if none).
 * @param holeCards - The player's hole cards.
 * @param communityCards - The shared board cards.
 * @returns Sets of hole and board indices that belong to the low hand.
 */
export function lowCardIndexSets(
  lowBestHand: Card[] | undefined,
  holeCards: Card[],
  communityCards: Card[],
): LowCardIndexSets {
  const loHoleSet = new Set<number>();
  const loBoardSet = new Set<number>();
  if (!lowBestHand || lowBestHand.length === 0) {
    return { loHoleSet, loBoardSet };
  }
  for (const low of lowBestHand) {
    const h = holeCards.findIndex((card) => sameCard(card, low));
    if (h >= 0) {
      loHoleSet.add(h);
      continue;
    }
    const b = communityCards.findIndex((card) => sameCard(card, low));
    if (b >= 0) loBoardSet.add(b);
  }
  return { loHoleSet, loBoardSet };
}
