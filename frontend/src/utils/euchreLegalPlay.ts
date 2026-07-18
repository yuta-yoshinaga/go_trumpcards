import type { Card } from '../types/card';
import { sameColorSuit } from './euchreBower';

/** Card design string → Euchre suit number (matches the domain's CardDesign constants). */
const DESIGN_TO_SUIT: Readonly<Record<string, number>> = {
  SPADE: 1,
  CLOVER: 2,
  HEART: 3,
  DIAMOND: 4,
  JOKER: 0,
};

/** Numeric value of a Jack, the only rank that can be a bower. */
const JACK_VALUE = 11;

/**
 * Returns the effective suit of a card under the given trump, mirroring the
 * domain's `Euchre.effectiveSuit`: the left bower (a Jack of the same-color
 * suit as trump) counts as the trump suit; every other card keeps its printed
 * suit. When `trumpSuit <= 0` (no trump set) the printed suit is always used.
 */
export function euchreEffectiveSuit(card: Card, trumpSuit: number): number {
  const suit = DESIGN_TO_SUIT[card.design] ?? 0;
  if (trumpSuit > 0 && card.value === JACK_VALUE && suit === sameColorSuit(trumpSuit)) {
    return trumpSuit;
  }
  return suit;
}

/**
 * Computes the indices of the human hand that are legal to play, mirroring the
 * domain's `Euchre.validatePlay` follow-suit rule: when leading (no lead card),
 * every card is legal; otherwise the player must follow the lead's effective
 * suit if able, so only matching cards are legal — but if the player holds none
 * of that suit, every card becomes legal. Effective suit accounts for the left
 * bower playing as trump.
 */
export function euchreLegalPlayIndices(hand: Card[], leadCard: Card | null | undefined, trumpSuit: number): number[] {
  const all = hand.map((_, i) => i);
  if (!leadCard) return all;
  const leadSuit = euchreEffectiveSuit(leadCard, trumpSuit);
  const following = all.filter((i) => euchreEffectiveSuit(hand[i], trumpSuit) === leadSuit);
  return following.length > 0 ? following : all;
}
