import type { Card } from '../types/card';

/** A joker's point value in Kalooki. Mirrors `KalookiJokerValue`. */
export const KALOOKI_JOKER_VALUE = 15;

/**
 * Point value of a single card: joker and ace 15, ten and above 10, otherwise
 * face value. Mirrors `kalookiCardValue` in `internal/domain/Kalooki.go`.
 * @param card - The card to score.
 * @returns Its point value.
 */
export function kalookiCardValue(card: Card): number {
  if (card.design === 'JOKER') return KALOOKI_JOKER_VALUE;
  if (card.value === 1) return 15;
  if (card.value >= 10) return 10;
  return card.value;
}

/**
 * Point value of one meld: the sum of its cards, multiplied by 1.5 and floored
 * when it contains a joker. Mirrors `kalookiMeldValue`.
 * @param cards - The cards forming the meld.
 * @returns The meld's points.
 */
export function kalookiMeldValue(cards: readonly Card[]): number {
  const base = cards.reduce((sum, c) => sum + kalookiCardValue(c), 0);
  const hasJoker = cards.some((c) => c.design === 'JOKER');
  // Integer arithmetic, floored, exactly as the domain does it.
  return hasJoker ? Math.floor((base * 3) / 2) : base;
}

/**
 * Total points of the staged meld groups, which is what an unopened player's
 * first meld is measured against.
 *
 * The joker bonus applies per meld, so the total is not simply the sum of the
 * card values — which is why the player cannot reliably do this in their head
 * and the readout is worth showing (#4839).
 * @param groups - Staged melds, each a list of cards.
 * @returns The combined points.
 */
export function kalookiOpeningPoints(groups: readonly (readonly Card[])[]): number {
  return groups.reduce((sum, g) => sum + kalookiMeldValue(g), 0);
}
