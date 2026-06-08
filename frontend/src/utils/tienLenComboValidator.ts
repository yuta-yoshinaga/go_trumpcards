import type { Card } from '../types/card';

/** Tien Len rank strength: 2 is highest (12), then A (11), then 3..K = 0..10. */
function valueStrength(v: number): number {
  if (v === 2) return 12;
  if (v === 1) return 11;
  return v - 3;
}

function allSameValue(cards: Card[]): boolean {
  return cards.length > 0 && cards.every((c) => c.value === cards[0].value);
}

/** A run of consecutive ranks (callers pass 3+ cards); 2 cannot appear, suits may be mixed. */
function checkStraight(cards: Card[]): boolean {
  if (cards.some((c) => c.value === 2)) return false;
  const strengths = cards.map((c) => valueStrength(c.value)).sort((a, b) => a - b);
  for (let i = 1; i < strengths.length; i++) {
    if (strengths[i] !== strengths[i - 1] + 1) return false;
  }
  return true;
}

/** Three consecutive pairs (e.g. 4-4-5-5-6-6); 2 cannot appear. */
function checkThreePairRun(cards: Card[]): boolean {
  if (cards.length !== 6) return false;
  if (cards.some((c) => c.value === 2)) return false;
  const freq = new Map<number, number>();
  for (const c of cards) freq.set(c.value, (freq.get(c.value) ?? 0) + 1);
  if (freq.size !== 3) return false;
  const strengths: number[] = [];
  for (const [v, cnt] of freq) {
    if (cnt !== 2) return false;
    strengths.push(valueStrength(v));
  }
  strengths.sort((a, b) => a - b);
  return strengths[1] === strengths[0] + 1 && strengths[2] === strengths[1] + 1;
}

/** The Tien Len play categories. */
export type TienLenCombo = 'single' | 'pair' | 'triple' | 'straight' | 'threePairRun' | 'fourOfAKind' | 'invalid';

/**
 * Classifies a selection of cards into its Tien Len play type, mirroring the Go
 * domain `tienLenClassifyPlay`.
 */
export function classifyTienLenCombo(cards: Card[]): TienLenCombo {
  const n = cards.length;
  if (n === 1) return 'single';
  if (n === 2) return allSameValue(cards) ? 'pair' : 'invalid';
  if (n === 3) return allSameValue(cards) ? 'triple' : checkStraight(cards) ? 'straight' : 'invalid';
  if (n === 4) return allSameValue(cards) ? 'fourOfAKind' : checkStraight(cards) ? 'straight' : 'invalid';
  if (n === 6) return checkThreePairRun(cards) ? 'threePairRun' : checkStraight(cards) ? 'straight' : 'invalid';
  return checkStraight(cards) ? 'straight' : 'invalid';
}

/** True if the selected cards form any legal Tien Len combination. */
export function isValidTienLenCombo(cards: Card[]): boolean {
  return cards.length > 0 && classifyTienLenCombo(cards) !== 'invalid';
}
