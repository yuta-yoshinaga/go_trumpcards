import type { Card, CardDesign } from '../types/card';

/**
 * Lexicographic compare of two descending-sorted card-value arrays.
 * Returns >0 if `a` ranks higher, <0 if `b` does, 0 if every position ties.
 * Used as High Card Flush's own showdown tiebreaker (compare 1st card, then
 * 2nd, then 3rd…).
 */
function compareDescendingValues(a: readonly number[], b: readonly number[]): number {
  const len = Math.min(a.length, b.length);
  for (let i = 0; i < len; i++) {
    if (a[i] !== b[i]) return a[i] - b[i];
  }
  return 0;
}

/**
 * Returns the suit (design) that forms the longest flush in a High Card Flush
 * hand, or null when the hand is empty. Tie-breaker matches the game's own
 * showdown rule: when two suits share the same count, the suit whose
 * descending-sorted values beat the other one lexicographically wins (compare
 * top card, then 2nd, then 3rd…). This stays deterministic even when the top
 * cards also tie — Map iteration order is never relied on.
 */
export function longestFlushSuit(hand: readonly Card[]): CardDesign | null {
  if (hand.length === 0) return null;
  const groups = new Map<CardDesign, number[]>();
  for (const c of hand) {
    const arr = groups.get(c.design);
    if (arr === undefined) {
      groups.set(c.design, [c.value]);
    } else {
      arr.push(c.value);
    }
  }
  let bestSuit: CardDesign | null = null;
  let bestVals: number[] = [];
  for (const [suit, vals] of groups) {
    vals.sort((a, b) => b - a);
    if (bestSuit === null || vals.length > bestVals.length) {
      bestSuit = suit;
      bestVals = vals;
      continue;
    }
    if (vals.length === bestVals.length && compareDescendingValues(vals, bestVals) > 0) {
      bestSuit = suit;
      bestVals = vals;
    }
  }
  return bestSuit;
}
