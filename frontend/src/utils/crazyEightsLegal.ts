import type { Card } from '../types/card';

/** Card design string → Crazy Eights suit number (mirrors `CrazyEightsSuit`). */
const DESIGN_TO_SUIT: Record<string, number> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

/**
 * Whether `card` may be legally played onto the current discard pile.
 *
 * Crazy Eights rules: an 8 is wild and always playable. After an 8 is played
 * the active suit is `chosenSuit` (1-4); otherwise a card must match the
 * discard top's suit or rank.
 *
 * @param card - The candidate card from the player's hand.
 * @param discardTop - The current top of the discard pile (`null` before any play).
 * @param chosenSuit - The suit declared by the last 8 (0 when none is active).
 * @returns `true` if the card can be legally played.
 */
export function isCrazyEightsLegalPlay(card: Card, discardTop: Card | null, chosenSuit: number): boolean {
  if (card.value === 8) return true; // eights are wild
  if (chosenSuit > 0) return (DESIGN_TO_SUIT[card.design] ?? 0) === chosenSuit;
  if (!discardTop) return true; // opening play
  return card.design === discardTop.design || card.value === discardTop.value;
}
