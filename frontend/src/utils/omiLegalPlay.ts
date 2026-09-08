import type { Card } from '../types/card';

/** Card design string → Omi suit number (matches the domain's CardDesign constants). */
const DESIGN_TO_SUIT: Readonly<Record<string, number>> = {
  SPADE: 1,
  CLOVER: 2,
  HEART: 3,
  DIAMOND: 4,
  JOKER: 0,
};

/**
 * Computes the indices of the human hand that are legal to play, following the
 * Omi follow-suit rule: when leading (no lead card), every card is legal;
 * otherwise the player must follow the lead suit if able — if the player holds
 * none of that suit, every card becomes legal.
 *
 * Unlike Euchre, Omi has no bower (jack promotion) mechanic, so each card is
 * evaluated purely on its printed suit.
 */
export function omiLegalPlayIndices(hand: Card[], leadCard: Card | null | undefined, _trumpSuit: number): number[] {
  const all = hand.map((_, i) => i);
  if (!leadCard) return all;
  const leadSuit = DESIGN_TO_SUIT[leadCard.design] ?? 0;
  const following = all.filter((i) => (DESIGN_TO_SUIT[hand[i].design] ?? 0) === leadSuit);
  return following.length > 0 ? following : all;
}
