import type { Card } from '../types/card';

/** Display sort modes shared by card game hands. */
export type HandSortMode = 'original' | 'rank' | 'suit';

/** Suit grouping order for display (♠♥♦♣), jokers always grouped last. */
const SUIT_ORDER: Record<string, number> = { SPADE: 0, HEART: 1, DIAMOND: 2, CLOVER: 3, JOKER: 4 };

const suitRank = (c: Card): number => SUIT_ORDER[c.design] ?? 5;

/**
 * Rank key used when ordering by value. Jokers (`design: 'JOKER'`, value 0) are
 * wildcards, so they sort to the very end rather than before the Ace.
 */
const valueRank = (c: Card): number => (c.design === 'JOKER' ? 99 : c.value);

/**
 * Returns the hand reordered for display under `mode`, each entry carrying its
 * **original** index so selection / meld / discard commands keep referencing the
 * server-dealt position (sorting is display-only and never invalidates the
 * indices sent to the backend).
 *
 * - `original`: server-dealt order (identity, indices 0..n-1).
 * - `rank`: by rank ascending (A,2,3…K), suit as tiebreak, jokers last.
 * - `suit`: grouped by suit (♠♥♦♣), rank within each suit, jokers last.
 */
export function sortedHandForDisplay(cards: readonly Card[], mode: HandSortMode): { card: Card; index: number }[] {
  const items = cards.map((card, index) => ({ card, index }));
  if (mode === 'original') return items;
  const comparators: Record<'rank' | 'suit', (a: Card, b: Card) => number> = {
    rank: (a, b) => valueRank(a) - valueRank(b) || suitRank(a) - suitRank(b),
    suit: (a, b) => suitRank(a) - suitRank(b) || valueRank(a) - valueRank(b),
  };
  const cmp = comparators[mode];
  // Stable sort keeps equal cards in their original relative order.
  return items.sort((a, b) => cmp(a.card, b.card) || a.index - b.index);
}
