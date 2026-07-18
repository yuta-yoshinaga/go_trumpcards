import type { Card } from '../types/card';

/** Fixed number of draw rounds per 2-7 Triple Draw hand. Mirrors the backend
 * constant `domain.DeuceToSevenMaxDraws` so the draw badge can show the cap. */
export const DEUCE_TO_SEVEN_MAX_DRAWS = 3;

/** Returns a card's value with the Ace counted as 14 (high), matching the 2-7
 * rule that the Ace never plays low. */
function highValue(c: Card): number {
  return c.value === 1 ? 14 : c.value;
}

/** Returns true when the 5 sorted high-values form a straight under 2-7 rules:
 * the Ace is always high, so A-2-3-4-5 is NOT a straight (A-10-J-Q-K is). */
function isDeuceStraight(sortedHigh: readonly number[]): boolean {
  if (sortedHigh.length !== 5) return false;
  for (let i = 1; i < sortedHigh.length; i++) {
    if (sortedHigh[i] !== sortedHigh[i - 1] + 1) return false;
  }
  return true;
}

/**
 * Returns true when a 5-card 2-7 hand is a "made pat low" — no pair, no
 * straight, no flush, and an 8-or-better high card (the strongest practical
 * shape, e.g. 8-6-4-3-2 or the nut 7-5-4-3-2). Drawing can only weaken it, so
 * the UI surfaces a stand-pat suggestion when this returns true.
 */
export function isMadePatLow(cards: ReadonlyArray<Card>): boolean {
  if (cards.length !== 5) return false;
  const ranks = new Set<number>();
  const suits = new Set<string>();
  const highs: number[] = [];
  for (const c of cards) {
    if (c.design === 'JOKER') return false;
    if (ranks.has(c.value)) return false; // any pair disqualifies
    ranks.add(c.value);
    suits.add(c.design);
    highs.push(highValue(c));
  }
  if (suits.size === 1) return false; // flush
  highs.sort((a, b) => a - b);
  if (isDeuceStraight(highs)) return false; // straight
  return highs[highs.length - 1] <= 8; // 8-or-better high card
}

/**
 * Indices of the cards worth keeping toward the best 2-7 low draw. A completed
 * pat low (see {@link isMadePatLow}) keeps all five cards. Otherwise the "core"
 * is the set of distinct 8-or-better ranks (the low targets): a pair's higher
 * duplicate and any 9-or-higher card are treated as draw candidates and left
 * out. Used to lift the keepers / dim the discards during the draw phase.
 */
export function deuceToSevenBestIndices(cards: ReadonlyArray<Card>): number[] {
  if (cards.length === 0) return [];
  // A made low can only be weakened by drawing — keep everything.
  if (isMadePatLow(cards)) return cards.map((_, i) => i);
  // Visit lowest-first so the kept copy of a pair is its (equal) lowest card.
  const order = cards
    .map((c, i) => ({ i, hv: highValue(c), rank: c.value, joker: c.design === 'JOKER' }))
    .sort((a, b) => a.hv - b.hv);
  const seenRank = new Set<number>();
  const keep: number[] = [];
  for (const { i, hv, rank, joker } of order) {
    if (joker || hv > 8 || seenRank.has(rank)) continue; // high card / pair dup → draw candidate
    seenRank.add(rank);
    keep.push(i);
  }
  return keep.sort((a, b) => a - b);
}
