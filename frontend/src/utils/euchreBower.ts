import type { Card } from '../types/card';

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
 * Returns the same-color suit for a Euchre trump suit: ♠(1)↔♣(2) (black),
 * ♥(3)↔♦(4) (red). Any other input is returned unchanged.
 */
export function sameColorSuit(suit: number): number {
  switch (suit) {
    case 1:
      return 2;
    case 2:
      return 1;
    case 3:
      return 4;
    case 4:
      return 3;
    default:
      return suit;
  }
}

/** The bower role of a card relative to the current trump, or `null` when it is neither. */
export type BowerRole = 'right' | 'left' | null;

/**
 * Determines whether a card is the right bower (Jack of the trump suit) or the
 * left bower (Jack of the same-color suit as trump) for the given trump suit.
 *
 * Returns `null` when no trump is set (`trumpSuit <= 0`) or the card is not a bower.
 */
export function bowerRole(card: Card, trumpSuit: number): BowerRole {
  if (trumpSuit <= 0 || card.value !== JACK_VALUE) return null;
  const cardSuit = DESIGN_TO_SUIT[card.design] ?? 0;
  if (cardSuit === trumpSuit) return 'right';
  if (cardSuit === sameColorSuit(trumpSuit)) return 'left';
  return null;
}
