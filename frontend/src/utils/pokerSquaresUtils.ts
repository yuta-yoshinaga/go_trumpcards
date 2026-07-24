import type { Card } from '../types/card';

/** Hand rank constants used by Poker Squares — mirror the backend rank ints. */
export const PokerHand = {
  HighCard: 0,
  OnePair: 1,
  TwoPair: 2,
  ThreeOfAKind: 3,
  Straight: 4,
  Flush: 5,
  FullHouse: 6,
  FourOfAKind: 7,
  StraightFlush: 8,
  RoyalFlush: 9,
} as const;

export type PokerHandRank = (typeof PokerHand)[keyof typeof PokerHand];

/** American Poker Squares scoring (mirrors `pokerSquaresScoreTable` in Go). */
const SCORE_TABLE: Record<PokerHandRank, number> = {
  [PokerHand.HighCard]: 0,
  [PokerHand.OnePair]: 2,
  [PokerHand.TwoPair]: 5,
  [PokerHand.ThreeOfAKind]: 10,
  [PokerHand.Straight]: 15,
  [PokerHand.Flush]: 20,
  [PokerHand.FullHouse]: 25,
  [PokerHand.FourOfAKind]: 50,
  [PokerHand.StraightFlush]: 75,
  [PokerHand.RoyalFlush]: 100,
};

/** Return the Poker Squares score for a hand rank. */
export function pokerSquaresRankToScore(rank: PokerHandRank): number {
  return SCORE_TABLE[rank];
}

/** i18n key suffix for each hand rank, indexed by `PokerHandRank` value. */
const POKER_HAND_KEYS = [
  'highCard',
  'onePair',
  'twoPair',
  'threeOfAKind',
  'straight',
  'flush',
  'fullHouse',
  'fourOfAKind',
  'straightFlush',
  'royalFlush',
] as const;

/** i18n key suffix for a hand rank. */
export function pokerHandKey(rank: PokerHandRank): string {
  return POKER_HAND_KEYS[rank];
}

/**
 * Evaluate a 5-card poker hand. Returns `null` for any non-5-card input.
 * Aces are low (value 1) when forming an A-2-3-4-5 straight and act as the
 * high anchor for 10-J-Q-K-A (mirroring the Go backend's behaviour).
 */
export function evaluateFiveCardHand(cards: readonly Card[]): PokerHandRank | null {
  if (cards.length !== 5) return null;

  const values = cards.map((c) => c.value).sort((a, b) => a - b);
  const designs = cards.map((c) => c.design);
  const counts = countByValue(values);
  const groupSizes = Object.values(counts).sort((a, b) => b - a);
  const isFlush = designs.every((d) => d === designs[0]);
  const isStraight = checkStraight(values);
  const isRoyal = isStraight && values[0] === 1 && values[4] === 13 && values[1] === 10;

  if (isFlush && isRoyal) return PokerHand.RoyalFlush;
  if (isFlush && isStraight) return PokerHand.StraightFlush;
  if (groupSizes[0] === 4) return PokerHand.FourOfAKind;
  if (groupSizes[0] === 3 && groupSizes[1] === 2) return PokerHand.FullHouse;
  if (isFlush) return PokerHand.Flush;
  if (isStraight) return PokerHand.Straight;
  if (groupSizes[0] === 3) return PokerHand.ThreeOfAKind;
  if (groupSizes[0] === 2 && groupSizes[1] === 2) return PokerHand.TwoPair;
  if (groupSizes[0] === 2) return PokerHand.OnePair;
  return PokerHand.HighCard;
}

/**
 * Evaluate the best *made* hand category for a partial (1-4 card) line.
 *
 * Only value multiples (pairs, trips, quads) can be definitively "made" before
 * a line is full, so straights, flushes, and full houses are intentionally
 * excluded — they are only realised once all five cards are present, where
 * {@link evaluateFiveCardHand} takes over. Returns `null` for an empty/full
 * line or when nothing better than a high card is formed, letting callers keep
 * the early-game board free of clutter while still hinting which lines are
 * developing a pair or better.
 */
export function evaluatePartialHand(cards: readonly Card[]): PokerHandRank | null {
  if (cards.length === 0 || cards.length >= 5) return null;
  const counts = countByValue(cards.map((c) => c.value));
  const groupSizes = Object.values(counts).sort((a, b) => b - a);
  if (groupSizes[0] === 4) return PokerHand.FourOfAKind;
  if (groupSizes[0] === 3) return PokerHand.ThreeOfAKind;
  if (groupSizes[0] === 2 && groupSizes[1] === 2) return PokerHand.TwoPair;
  if (groupSizes[0] === 2) return PokerHand.OnePair;
  return null;
}

/**
 * Evaluate the best 5-card poker hand reachable from up to seven cards.
 *
 * Composes {@link evaluateFiveCardHand} over every C(n, 5) combination once at
 * least five cards are present (n is tiny — at most C(7,5)=21 combinations — so
 * the brute-force scan is cheap). With one to four cards it falls back to
 * {@link evaluatePartialHand}, defaulting to {@link PokerHand.HighCard} when no
 * multiple is made yet, so a live readout always has a name to show. Returns
 * `null` only for an empty card list.
 */
export function evaluateBestHand(cards: readonly Card[]): PokerHandRank | null {
  if (cards.length === 0) return null;
  if (cards.length < 5) return evaluatePartialHand(cards) ?? PokerHand.HighCard;
  let best: PokerHandRank | null = null;
  for (const combo of fiveCardCombinations(cards)) {
    const rank = evaluateFiveCardHand(combo);
    if (rank !== null && (best === null || rank > best)) best = rank;
  }
  return best;
}

/** Yield every 5-card combination of the given cards (n is small: n ≤ 7). */
function* fiveCardCombinations(cards: readonly Card[]): Generator<Card[]> {
  const n = cards.length;
  const idx = [0, 1, 2, 3, 4];
  while (true) {
    yield idx.map((i) => cards[i]);
    let k = 4;
    while (k >= 0 && idx[k] === n - 5 + k) k--;
    if (k < 0) return;
    idx[k]++;
    for (let j = k + 1; j < 5; j++) idx[j] = idx[j - 1] + 1;
  }
}

function countByValue(sortedValues: readonly number[]): Record<number, number> {
  const out: Record<number, number> = {};
  for (const v of sortedValues) out[v] = (out[v] ?? 0) + 1;
  return out;
}

function checkStraight(sortedValues: readonly number[]): boolean {
  // Five distinct sequential ranks, or the A-2-3-4-5 wheel.
  if (new Set(sortedValues).size !== 5) return false;
  if (sortedValues[4] - sortedValues[0] === 4) return true;
  // 10-J-Q-K-A becomes [1, 10, 11, 12, 13] after sort. Treat that as a straight.
  return sortedValues[0] === 1 && sortedValues[1] === 10 && sortedValues[4] === 13;
}
