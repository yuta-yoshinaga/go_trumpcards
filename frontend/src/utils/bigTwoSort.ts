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
