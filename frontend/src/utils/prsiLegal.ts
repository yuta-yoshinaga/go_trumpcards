import type { Card } from '../types/card';

/**
 * Whether `card` may be legally played onto the current Prší discard pile.
 *
 * Prší rules (Czech Crazy Eights): a card must match the discard top's suit or
 * rank. There is no wild card. When a 7-stack penalty is active
 * (`penaltyDrawCount > 0`), the only legal response is to stack another 7;
 * otherwise the player must draw the penalty.
 *
 * @param card - The candidate card from the player's hand.
 * @param discardTop - The current top of the discard pile (`null` before any play).
 * @param penaltyDrawCount - The accumulated 7-stack penalty (0 when none active).
 * @returns `true` if the card can be legally played.
 */
export function isPrsiLegalPlay(card: Card, discardTop: Card | null, penaltyDrawCount: number): boolean {
  if (penaltyDrawCount > 0) return card.value === 7; // under penalty only a 7 stacks
  if (!discardTop) return true; // opening play
  return card.design === discardTop.design || card.value === discardTop.value;
}
