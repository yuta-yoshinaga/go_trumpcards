import type { Card, CardDesign } from '../types/card';

/** Numeric score a single card contributes in Fifty-One: A=11, J/Q/K=10, 2-10=face value. */
export function fiftyOneCardScore(value: number): number {
  if (value === 1) return 11;
  if (value >= 11) return 10;
  return value;
}

/** Per-suit total scores for a Fifty-One hand. Joker cards are ignored. */
export type FiftyOneSuitScoreMap = Record<Exclude<CardDesign, 'JOKER'>, number>;

/** Aggregate per-suit totals for a hand. */
export function fiftyOneSuitScores(cards: Card[]): FiftyOneSuitScoreMap {
  const scores: FiftyOneSuitScoreMap = { SPADE: 0, CLOVER: 0, HEART: 0, DIAMOND: 0 };
  for (const c of cards) {
    if (c.design === 'JOKER') continue;
    scores[c.design] += fiftyOneCardScore(c.value);
  }
  return scores;
}

/** Suit (excluding JOKER) with the highest total, breaking ties in fixed order S/C/H/D. */
export function fiftyOneBestSuit(scores: FiftyOneSuitScoreMap): Exclude<CardDesign, 'JOKER'> {
  const order: Array<Exclude<CardDesign, 'JOKER'>> = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'];
  let best: Exclude<CardDesign, 'JOKER'> = 'SPADE';
  let bestScore = -1;
  for (const d of order) {
    if (scores[d] > bestScore) {
      best = d;
      bestScore = scores[d];
    }
  }
  return best;
}
