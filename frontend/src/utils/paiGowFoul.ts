import type { Card } from '../types/card';

/** Numeric rank used for Pai Gow comparisons: Ace = 14, J/Q/K = 11/12/13, otherwise face value. */
function paiGowValue(card: Card): number {
  if (card.value === 1) return 14;
  return card.value;
}

function isJoker(card: Card): boolean {
  return card.design === 'JOKER';
}

/** Detect a 5-card straight ignoring suits (handles A-low 1-5 and A-high 10-A). */
function isStraightValues(vals: number[]): boolean {
  const sorted = [...vals].sort((a, b) => a - b);
  // A-low: 14,2,3,4,5 → after sort 2,3,4,5,14
  if (sorted[0] === 2 && sorted[1] === 3 && sorted[2] === 4 && sorted[3] === 5 && sorted[4] === 14) return true;
  for (let i = 1; i < sorted.length; i++) {
    if (sorted[i] - sorted[i - 1] !== 1) return false;
  }
  return true;
}

interface HighHandStrength {
  /** True if rank is two pair, trips, straight, flush, or better — always beats any 2-card low hand. */
  isRank2Plus: boolean;
  /** The value of the single pair, or 0 if the high hand is not exactly one pair. */
  pairValue: number;
  /** High hand values sorted descending, used for high-card comparison. */
  highVals: number[];
}

/** Compute the strength category of a 5-card high hand for foul detection. */
function getHighHandStrength(highHand: Card[]): HighHandStrength {
  const counts = new Map<number, number>();
  let pairs = 0;
  let highPairVal = 0;
  let hasTripsOrBetter = false;

  for (const card of highHand) {
    const v = paiGowValue(card);
    const count = (counts.get(v) ?? 0) + 1;
    counts.set(v, count);
    if (count === 2) {
      pairs++;
      if (v > highPairVal) highPairVal = v;
    }
    if (count >= 3) hasTripsOrBetter = true;
  }

  const highVals = highHand.map(paiGowValue).sort((a, b) => b - a);
  const isStraightOrFlush = isStraightValues(highVals) || highHand.every((c) => c.design === highHand[0].design);

  return {
    isRank2Plus: isStraightOrFlush || hasTripsOrBetter || pairs >= 2,
    pairValue: pairs === 1 ? highPairVal : 0,
    highVals,
  };
}

/** Computed foul state for a Pai Gow split. */
export interface PaiGowFoulResult {
  /** True when the split is illegal (low hand outranks high). */
  isFoul: boolean;
}

/** Detect whether the 2-card low hand outranks the 5-card high hand. Skips evaluation if a joker is present. */
export function paiGowFoulCheck(playerCards: Card[], lowIndices: readonly number[]): PaiGowFoulResult {
  if (lowIndices.length !== 2 || playerCards.length !== 7) return { isFoul: false };
  if (playerCards.some(isJoker)) return { isFoul: false }; // conservative — backend handles
  const [i0, i1] = lowIndices;
  if (i0 === i1 || i0 < 0 || i1 < 0 || i0 >= 7 || i1 >= 7) return { isFoul: false };
  const lowHand = [playerCards[i0], playerCards[i1]];
  const highHand = playerCards.filter((_, idx) => idx !== i0 && idx !== i1);

  const lowVals = lowHand.map(paiGowValue).sort((a, b) => b - a);
  const isLowPair = lowVals[0] === lowVals[1];
  const highInfo = getHighHandStrength(highHand);

  // Rank 2+ (two pair, trips, straight, flush, or better) always beats any 2-card hand.
  if (highInfo.isRank2Plus) return { isFoul: false };

  if (isLowPair) {
    // Low pair vs high card: foul (pair beats any high card hand).
    if (highInfo.pairValue === 0) return { isFoul: true };
    // Low pair vs high pair: foul only when the low pair outranks the high pair.
    return { isFoul: lowVals[0] > highInfo.pairValue };
  }

  // Low is high card. High hand has a pair (beats any high card): not a foul.
  if (highInfo.pairValue > 0) return { isFoul: false };

  // Both are high card: compare top card, then second card.
  if (lowVals[0] !== highInfo.highVals[0]) return { isFoul: lowVals[0] > highInfo.highVals[0] };
  return { isFoul: lowVals[1] > highInfo.highVals[1] };
}

/** Lexicographically compare two descending value arrays: positive when `a` outranks `b`. */
function compareDescValues(a: number[], b: number[]): number {
  for (let i = 0; i < a.length && i < b.length; i++) {
    if (a[i] !== b[i]) return a[i] - b[i];
  }
  return 0;
}

/**
 * Compute a "house way" style low-hand split: the pair of indices (into the 7-card
 * player hand) that forms the strongest legal 2-card low hand without fouling.
 *
 * It enumerates all C(7,2)=21 splits, keeps only the non-foul ones (via
 * {@link paiGowFoulCheck}), and picks the split that maximizes the low hand
 * (a pair beats a high card, then higher card values win) — the same primary
 * criterion the dealer's house-way algorithm uses.
 *
 * Returns `null` when the hand contains a joker (foul evaluation is unavailable, so
 * an auto-split cannot be guaranteed legal) or when the hand is not exactly 7 cards.
 */
export function paiGowAutoSplit(playerCards: Card[]): [number, number] | null {
  if (playerCards.length !== 7) return null;
  if (playerCards.some(isJoker)) return null;

  let best: { indices: [number, number]; rank: number; vals: number[] } | null = null;
  for (let i = 0; i < 7; i++) {
    for (let j = i + 1; j < 7; j++) {
      if (paiGowFoulCheck(playerCards, [i, j]).isFoul) continue;
      const vals = [paiGowValue(playerCards[i]), paiGowValue(playerCards[j])].sort((a, b) => b - a);
      const rank = vals[0] === vals[1] ? 1 : 0; // a pair outranks a high card
      if (best === null || rank > best.rank || (rank === best.rank && compareDescValues(vals, best.vals) > 0)) {
        best = { indices: [i, j], rank, vals };
      }
    }
  }
  return best?.indices ?? null;
}
