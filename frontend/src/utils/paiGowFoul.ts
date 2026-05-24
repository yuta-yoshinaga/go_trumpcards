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

/** True if the 5-card high hand is at least one pair OR a straight/flush (which beats any 2-card low). */
function highHandAtLeastPair(highHand: Card[]): boolean {
  if (highHand.length !== 5) return false;
  const counts = new Map<number, number>();
  for (const c of highHand) {
    const v = paiGowValue(c);
    counts.set(v, (counts.get(v) ?? 0) + 1);
  }
  for (const c of counts.values()) {
    if (c >= 2) return true;
  }
  const vals = highHand.map(paiGowValue);
  if (isStraightValues(vals)) return true;
  const firstDesign = highHand[0].design;
  if (highHand.every((c) => c.design === firstDesign)) return true;
  return false;
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

  if (highHandAtLeastPair(highHand)) return { isFoul: false };
  if (isLowPair) return { isFoul: true };
  const highVals = highHand.map(paiGowValue).sort((a, b) => b - a);
  return { isFoul: highVals[0] < lowVals[0] };
}
