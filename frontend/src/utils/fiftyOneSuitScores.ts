import type { Card, CardDesign } from '../types/card';

/**
 * Maximum possible score in Fifty-One (A+K+Q+J+10 of the same suit = 51).
 * `internal/domain/FiftyOnePlayer.go` の `FiftyOneMaxScore` と同期。片方だけ変えないこと。
 */
export const FIFTY_ONE_MAX_SCORE = 51;

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
