import type { Card } from '../types/card';

/**
 * Choose the best 5-card poker hand out of a 7-card set (2 hole + 5 community).
 * Returns the indices (into the input array) of the 5 cards that comprise the
 * best hand, or `null` if fewer than 5 cards are supplied.
 *
 * The evaluator scores each of the 21 5-card subsets and picks the highest.
 * Hand ordering uses a lexicographic tuple:
 *   [category, primary..., kickers...]
 * with Aces both high (14) and as the low end of A-2-3-4-5 (wheel).
 */
export function holdemBestFive(cards: readonly Card[]): number[] | null {
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
            const score = scoreFive(idx.map((i) => cards[i]));
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

function compareScores(a: readonly number[], b: readonly number[]): number {
  const len = Math.max(a.length, b.length);
  for (let i = 0; i < len; i += 1) {
    const av = a[i] ?? 0;
    const bv = b[i] ?? 0;
    if (av !== bv) return av - bv;
  }
  return 0;
}

function rankValue(card: Card): number {
  // Treat Aces as high (14); the straight check handles the wheel separately.
  return card.value === 1 ? 14 : card.value;
}

function scoreFive(hand: readonly Card[]): number[] {
  const ranks = hand.map(rankValue).sort((x, y) => y - x);
  const suits = hand.map((c) => c.design);
  const flush = suits.every((s) => s === suits[0]);
  const uniq = Array.from(new Set(ranks)).sort((x, y) => y - x);
  const straight = isStraight(uniq);

  // Group by rank, sort by (count desc, rank desc).
  const counts = new Map<number, number>();
  for (const r of ranks) counts.set(r, (counts.get(r) ?? 0) + 1);
  const grouped = Array.from(counts.entries()).sort((a, b) => b[1] - a[1] || b[0] - a[0]);

  if (flush && straight !== null) {
    if (straight === 14) return [9, 14, 13, 12, 11, 10];
    return [9, straight];
  }
  if (grouped[0][1] === 4) {
    const four = grouped[0][0];
    const kicker = grouped[1][0];
    return [8, four, kicker];
  }
  if (grouped[0][1] === 3 && grouped[1] && grouped[1][1] === 2) {
    return [7, grouped[0][0], grouped[1][0]];
  }
  if (flush) {
    return [6, ...ranks];
  }
  if (straight !== null) {
    return [5, straight];
  }
  if (grouped[0][1] === 3) {
    const three = grouped[0][0];
    const kickers = grouped.slice(1).map(([r]) => r);
    return [4, three, ...kickers];
  }
  if (grouped[0][1] === 2 && grouped[1] && grouped[1][1] === 2) {
    const [hi, lo] = grouped[0][0] > grouped[1][0] ? [grouped[0][0], grouped[1][0]] : [grouped[1][0], grouped[0][0]];
    const kicker = grouped[2]?.[0] ?? 0;
    return [3, hi, lo, kicker];
  }
  if (grouped[0][1] === 2) {
    const pair = grouped[0][0];
    const kickers = grouped.slice(1).map(([r]) => r);
    return [2, pair, ...kickers];
  }
  return [1, ...ranks];
}

/**
 * Return the high card of a 5-rank straight if present, otherwise null.
 * Aces as wheel (A-2-3-4-5 = high 5) are handled via the duplicate-low-1 pass.
 */
function isStraight(uniqSortedDesc: readonly number[]): number | null {
  if (uniqSortedDesc.length < 5) return null;
  // Regular: 5 consecutive.
  for (let i = 0; i + 4 < uniqSortedDesc.length; i += 1) {
    if (uniqSortedDesc[i] - uniqSortedDesc[i + 4] === 4) return uniqSortedDesc[i];
  }
  // Wheel: A-2-3-4-5 (Aces stored as 14 here).
  const withWheel = uniqSortedDesc.slice();
  if (
    withWheel[0] === 14 &&
    withWheel.includes(5) &&
    withWheel.includes(4) &&
    withWheel.includes(3) &&
    withWheel.includes(2)
  ) {
    return 5;
  }
  return null;
}
