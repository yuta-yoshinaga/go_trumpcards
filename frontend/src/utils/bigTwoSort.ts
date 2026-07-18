import type { Card } from '../types/card';

/**
 * Big Two value strength (2 is highest, then A, K…3). Mirrors
 * `bigTwoValueStrength` in internal/domain/BigTwoEval.go: 2→12, A→11, else value-3.
 */
export function bigTwoValueStrength(value: number): number {
  if (value === 2) return 12;
  if (value === 1) return 11; // Ace
  return value - 3;
}

/** Suit strength ♦<♣<♥<♠ (0-3), mirroring `bigTwoSuitStrength`. */
const SUIT_STRENGTH: Record<string, number> = { DIAMOND: 0, CLOVER: 1, HEART: 2, SPADE: 3 };

/** Combined Big Two card strength = valueStrength*4 + suitStrength (matches `BigTwoCardStrength`). */
export function bigTwoCardStrength(card: Card): number {
  return bigTwoValueStrength(card.value) * 4 + (SUIT_STRENGTH[card.design] ?? 0);
}

/** Hand sort modes offered in the UI. */
export type BigTwoSortMode = 'strength' | 'suit' | 'number';

/**
 * Returns the hand reordered for display under `mode`, each entry carrying its
 * **original** index so selection/play commands keep referencing the
 * server-dealt position (sorting never invalidates the selected indices).
 *
 * - `strength`: Big Two play strength (2 high), suit as tiebreak.
 * - `suit`: grouped by suit (♦♣♥♠), strength within each suit.
 * - `number`: natural rank order (A,2,3…K), suit as tiebreak.
 */
export function sortedBigTwoHand(cards: readonly Card[], mode: BigTwoSortMode): { card: Card; index: number }[] {
  const items = cards.map((card, index) => ({ card, index }));
  const suit = (c: Card) => SUIT_STRENGTH[c.design] ?? 0;
  const comparators: Record<BigTwoSortMode, (a: Card, b: Card) => number> = {
    strength: (a, b) => bigTwoCardStrength(a) - bigTwoCardStrength(b),
    suit: (a, b) => suit(a) - suit(b) || bigTwoValueStrength(a.value) - bigTwoValueStrength(b.value),
    number: (a, b) => a.value - b.value || suit(a) - suit(b),
  };
  const cmp = comparators[mode];
  return [...items].sort((a, b) => cmp(a.card, b.card));
}

/**
 * Checks whether 5 rank values (any order) form a Big Two straight.
 *
 * Mirrors `bigTwoCheckStraight` in `internal/domain/BigTwoEval.go`: a 2 can
 * never be part of a straight, 10-J-Q-K-A is the only Ace-high run, and the
 * Ace (value 1) never sits at the low end.
 */
function bigTwoIsStraight(values: readonly number[]): boolean {
  const sorted = [...values].sort((a, b) => a - b);
  if (sorted.includes(2)) return false;
  // 10-J-Q-K-A (Ace high) — sorted numerically this is [1,10,11,12,13].
  if (sorted[0] === 1 && sorted[1] === 10 && sorted[2] === 11 && sorted[3] === 12 && sorted[4] === 13) {
    return true;
  }
  if (sorted[0] === 1) return false; // Ace can only be the high end (handled above).
  for (let i = 1; i < 5; i++) {
    if (sorted[i] !== sorted[i - 1] + 1) return false;
  }
  return true;
}

/** Classifies a 5-card selection into its Big Two play type number (5-8), or 0 if invalid. */
function bigTwoClassify5(cards: readonly Card[]): number {
  const values = cards.map((c) => c.value);
  const isFlush = cards.every((c) => c.design === cards[0].design);
  const isStraight = bigTwoIsStraight(values);

  const freq = new Map<number, number>();
  for (const v of values) freq.set(v, (freq.get(v) ?? 0) + 1);
  const counts = [...freq.values()].sort((a, b) => b - a);

  if (isFlush && isStraight) return 8; // straight flush
  if (counts[0] === 4) return 7; // four of a kind (+ kicker)
  if (counts[0] === 3 && counts[1] === 2) return 6; // full house
  if (isFlush) return 5; // flush
  if (isStraight) return 4; // straight
  return 0;
}

/**
 * Classifies a selection of cards into its Big Two play type number, mirroring
 * `bigTwoClassifyPlay` in `internal/domain/BigTwoEval.go`.
 *
 * Returns 1=single, 2=pair, 3=triple, 4=straight, 5=flush, 6=full house,
 * 7=four of a kind, 8=straight flush, or 0 for any invalid combination
 * (including empty, 4-card, and 6+-card selections). This classifies the
 * combination shape only — it does not check play legality against the table.
 */
export function classifyBigTwoPlay(cards: readonly Card[]): number {
  switch (cards.length) {
    case 1:
      return 1;
    case 2:
      return cards[0].value === cards[1].value ? 2 : 0;
    case 3:
      return cards[0].value === cards[1].value && cards[1].value === cards[2].value ? 3 : 0;
    case 5:
      return bigTwoClassify5(cards);
    default:
      return 0;
  }
}

/** i18n key suffix for a Big Two table play type (1-8), or `null` when none is on the table. */
export function bigTwoPlayTypeKey(playType: number): string | null {
  switch (playType) {
    case 1:
      return 'single';
    case 2:
      return 'pair';
    case 3:
      return 'triple';
    case 4:
      return 'straight';
    case 5:
      return 'flush';
    case 6:
      return 'fullHouse';
    case 7:
      return 'fourOfAKind';
    case 8:
      return 'straightFlush';
    default:
      return null;
  }
}
