import type { Card } from '../types/card';
import { compareScores } from './holdemBestFive';

/**
 * Short Deck (6+ Hold'em) hand scoring.
 *
 * Short Deck is played with a 36-card deck (A, 6–K) and differs from standard
 * poker in exactly two ways, both mirrored from
 * `internal/domain/shortdeck_hand_eval.go`:
 *
 * 1. A flush beats a full house (`ShortDeckHandFlush` = 6 > `ShortDeckHandFullHouse` = 5).
 * 2. A-6-7-8-9 is the wheel — the lowest straight — since 2–5 are not in the deck.
 *
 * Reusing the standard evaluator would pick a full house over a flush and would
 * not see the wheel at all, so a highlight built on it could mark five cards
 * that are not the hand the server actually scored.
 */

/** Ace counts as 14 unless it is the low end of the A-6-7-8-9 wheel. */
function rankValue(card: Card): number {
  return card.value === 1 ? 14 : card.value;
}

/**
 * High card of the straight these ranks form, or null when they do not form one.
 * Returns 9 for the A-6-7-8-9 wheel, where the ace plays low.
 * @param descendingUnique - Distinct ranks, highest first, aces as 14.
 * @returns The straight's high rank, or null.
 */
export function shortDeckStraightHigh(descendingUnique: readonly number[]): number | null {
  if (descendingUnique.length !== 5) return null;
  // A-6-7-8-9: [14, 9, 8, 7, 6] with the ace playing below the six.
  if (
    descendingUnique[0] === 14 &&
    descendingUnique[1] === 9 &&
    descendingUnique[2] === 8 &&
    descendingUnique[3] === 7 &&
    descendingUnique[4] === 6
  ) {
    return 9;
  }
  for (let i = 1; i < descendingUnique.length; i += 1) {
    if ((descendingUnique[i] ?? 0) !== (descendingUnique[i - 1] ?? 0) - 1) return null;
  }
  return descendingUnique[0] ?? null;
}

/**
 * Score a 5-card Short Deck hand as a lexicographic tuple `[category, …]`,
 * higher being stronger. Categories follow the Short Deck order, so a flush
 * outranks a full house.
 * @param hand - Exactly five cards.
 * @returns The comparable score tuple.
 */
export function scoreFiveShortDeck(hand: readonly Card[]): number[] {
  const ranks = hand.map(rankValue).sort((x, y) => y - x);
  const suits = hand.map((c) => c.design);
  const flush = suits.every((s) => s === suits[0]);
  const uniq = Array.from(new Set(ranks)).sort((x, y) => y - x);
  const straight = shortDeckStraightHigh(uniq);

  const counts = new Map<number, number>();
  for (const r of ranks) counts.set(r, (counts.get(r) ?? 0) + 1);
  const grouped = Array.from(counts.entries()).sort((a, b) => b[1] - a[1] || b[0] - a[0]);
  const top = grouped[0] ?? [0, 0];
  const second = grouped[1];

  if (flush && straight !== null) return [9, straight];
  if (top[1] === 4) return [8, top[0], second?.[0] ?? 0];
  // The Short Deck swap: flush above full house.
  if (flush) return [7, ...ranks];
  if (top[1] === 3 && second && second[1] === 2) return [6, top[0], second[0]];
  if (straight !== null) return [5, straight];
  if (top[1] === 3) return [4, top[0], ...grouped.slice(1).map(([r]) => r)];
  if (top[1] === 2 && second && second[1] === 2) {
    const [hi, lo] = top[0] > second[0] ? [top[0], second[0]] : [second[0], top[0]];
    return [3, hi, lo, grouped[2]?.[0] ?? 0];
  }
  if (top[1] === 2) return [2, top[0], ...grouped.slice(1).map(([r]) => r)];
  return [1, ...ranks];
}

/**
 * Indices of the five cards forming the best Short Deck hand in the supplied
 * set (2 hole + 5 community), or null with fewer than five cards.
 * @param cards - The candidate cards.
 * @returns Indices into `cards`, or null.
 */
export function shortDeckBestFive(cards: readonly Card[]): number[] | null {
  if (cards.length < 5) return null;
  let bestScore: number[] = [];
  let bestIdx: number[] = [];
  const n = cards.length;
  for (let a = 0; a < n - 4; a += 1) {
    for (let b = a + 1; b < n - 3; b += 1) {
      for (let c = b + 1; c < n - 2; c += 1) {
        for (let d = c + 1; d < n - 1; d += 1) {
          for (let e = d + 1; e < n; e += 1) {
            const idx = [a, b, c, d, e];
            const score = scoreFiveShortDeck(idx.map((i) => cards[i] as Card));
            if (compareScores(score, bestScore) > 0) {
              bestScore = score;
              bestIdx = idx;
            }
          }
        }
      }
    }
  }
  return bestIdx;
}
