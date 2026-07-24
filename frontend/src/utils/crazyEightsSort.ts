import type { Card } from '../types/card';

/** Hand sort modes offered in the Crazy Eights footer. */
export type CrazyEightsSortMode = 'original' | 'rank' | 'suit';

/** localStorage key persisting the player's chosen Crazy Eights hand-sort mode. */
export const CRAZYEIGHTS_SORT_STORAGE_KEY = 'crazyeights-sort-mode';

/** Suit grouping order for display (♠♥♦♣); the standard deck carries no jokers. */
const SUIT_ORDER: Record<string, number> = { SPADE: 0, HEART: 1, DIAMOND: 2, CLOVER: 3 };

const suitRank = (c: Card): number => SUIT_ORDER[c.design] ?? 4;

/**
 * Returns the hand reordered for display under `mode`, each entry carrying its
 * **original** index so selection / play commands keep referencing the
 * server-dealt position (sorting is display-only and never invalidates the
 * indices sent to the backend).
 *
 * - `original`: server-dealt order (identity, indices 0..n-1).
 * - `rank`: by rank ascending, suit as tiebreak.
 * - `suit`: grouped by suit (♠♥♦♣), rank within each suit.
 */
export function sortedCrazyEightsHand(
  cards: readonly Card[],
  mode: CrazyEightsSortMode,
): { card: Card; index: number }[] {
  const items = cards.map((card, index) => ({ card, index }));
  if (mode === 'original') return items;
  const comparators: Record<'rank' | 'suit', (a: Card, b: Card) => number> = {
    rank: (a, b) => a.value - b.value || suitRank(a) - suitRank(b),
    suit: (a, b) => suitRank(a) - suitRank(b) || a.value - b.value,
  };
  const cmp = comparators[mode];
  // Stable sort keeps equal cards in their original relative order.
  return items.sort((a, b) => cmp(a.card, b.card) || a.index - b.index);
}

/** Reads the persisted sort mode from localStorage, defaulting to `original`. */
export function loadCrazyEightsSortMode(): CrazyEightsSortMode {
  try {
    const v = localStorage.getItem(CRAZYEIGHTS_SORT_STORAGE_KEY);
    if (v === 'rank' || v === 'suit' || v === 'original') return v;
  } catch {
    // Ignore storage access errors (e.g. privacy mode) and fall back to default.
  }
  return 'original';
}

/** Persists the chosen sort mode to localStorage, ignoring storage failures. */
export function saveCrazyEightsSortMode(mode: CrazyEightsSortMode): void {
  try {
    localStorage.setItem(CRAZYEIGHTS_SORT_STORAGE_KEY, mode);
  } catch {
    // Ignore storage access errors (e.g. privacy mode).
  }
}
