import type { Card } from '../types/card';

/** Numeric suit index for each card design (1=♠ 2=♣ 3=♥ 4=♦; 0 for JOKER/unknown). */
const DESIGN_TO_SUIT: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

/** Same-color partner suit (♠↔♣, ♥↔♦), used to locate the Left Pedro. */
const SAME_COLOR: Readonly<Record<number, number>> = { 1: 2, 2: 1, 3: 4, 4: 3 };

/** Candidate trump suits, 1=♠ 2=♣ 3=♥ 4=♦. */
export const CINCH_TRUMP_SUITS = [1, 2, 3, 4] as const;

/**
 * Point value a card contributes toward the 14 deal points when `suit` is trump.
 *
 * Mirrors `cinchPointValue` in `internal/domain/Cinch.go`: the trump A/K/10/J
 * are worth 1 each ("High"/"King"/"Ten (Game)"/"Jack"), the Right Pedro (5 of
 * trump) is worth 5, and the Left Pedro (5 of the same-color off-suit) is worth
 * 5. Every other card is worth 0.
 */
export function cinchCardPoints(card: Card, suit: number): number {
  const cardSuit = DESIGN_TO_SUIT[card.design] ?? 0;
  if (cardSuit === 0) return 0;
  // Left Pedro: the 5 of the same-color off-suit counts as a trump point card.
  if (card.value === 5 && cardSuit === SAME_COLOR[suit]) return 5;
  if (cardSuit !== suit) return 0;
  switch (card.value) {
    case 1: // High (A)
    case 13: // King
    case 11: // Jack
    case 10: // Ten (Game)
      return 1;
    case 5: // Right Pedro
      return 5;
    default:
      return 0;
  }
}

/** A rough, hand-only estimate of Cinch bidding strength. */
export interface CinchBidStrength {
  /** Point-card points held in hand for each candidate trump suit (index 1-4). */
  pointsBySuit: Readonly<Record<number, number>>;
  /** Strongest candidate trump suit (1-4); the lowest suit index wins ties. */
  bestSuit: number;
  /** Points held in the strongest suit — the upper end of the guide range. */
  maxPoints: number;
  /** Points held in the weakest suit — the lower end of the guide range. */
  minPoints: number;
}

/**
 * Estimate how many of the 14 deal points the hand already holds, per candidate
 * trump suit. This is a rough guide only: holding a point card is not the same
 * as capturing it in play, and the estimate ignores trump length / control.
 * `maxPoints`/`minPoints` bracket the "depending on trump" range, and `bestSuit`
 * is the suit that maximizes held points.
 */
export function estimateCinchBidStrength(cards: Card[]): CinchBidStrength {
  const pointsBySuit: Record<number, number> = { 1: 0, 2: 0, 3: 0, 4: 0 };
  for (const suit of CINCH_TRUMP_SUITS) {
    let total = 0;
    for (const c of cards) total += cinchCardPoints(c, suit);
    pointsBySuit[suit] = total;
  }
  let bestSuit: number = CINCH_TRUMP_SUITS[0];
  for (const suit of CINCH_TRUMP_SUITS) {
    if (pointsBySuit[suit] > pointsBySuit[bestSuit]) bestSuit = suit;
  }
  const values = CINCH_TRUMP_SUITS.map((s) => pointsBySuit[s]);
  return {
    pointsBySuit,
    bestSuit,
    maxPoints: Math.max(...values),
    minPoints: Math.min(...values),
  };
}
