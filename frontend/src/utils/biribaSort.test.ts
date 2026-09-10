import { afterEach, describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  BIRIBA_SORT_STORAGE_KEY,
  type BiribaSortMode,
  loadBiribaSortMode,
  saveBiribaSortMode,
  sortedBiribaHand,
} from './biribaSort';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('sortedBiribaHand', () => {
  const hand = [c('DIAMOND', 5), c('SPADE', 3), c('JOKER', 0), c('HEART', 3), c('SPADE', 13)];

  it('original mode returns identity order with matching indices', () => {
    const out = sortedBiribaHand(hand, 'original');
    expect(out.map((o) => o.index)).toEqual([0, 1, 2, 3, 4]);
    expect(out.map((o) => o.card)).toEqual(hand);
  });

  it('keeps each card paired with its original index in every mode', () => {
    for (const mode of ['original', 'rank', 'suit'] as BiribaSortMode[]) {
      for (const { card, index } of sortedBiribaHand(hand, mode)) {
        expect(hand[index]).toBe(card);
      }
    }
  });

  it('rank mode orders by value ascending with the joker last', () => {
    const out = sortedBiribaHand(hand, 'rank');
    expect(out.map((o) => o.card.value)).toEqual([3, 3, 5, 13, 0]);
    expect(out[out.length - 1].card.design).toBe('JOKER');
  });

  it('suit mode groups by suit ♠♥♦♣ with the joker last', () => {
    const out = sortedBiribaHand(hand, 'suit');
    expect(out.map((o) => o.card.design)).toEqual(['SPADE', 'SPADE', 'HEART', 'DIAMOND', 'JOKER']);
    // Within SPADE, rank ascending: 3 before 13.
    expect(out[0].card.value).toBe(3);
    expect(out[1].card.value).toBe(13);
  });

  it('does not mutate the input array', () => {
    const snapshot = [...hand];
    sortedBiribaHand(hand, 'rank');
    expect(hand).toEqual(snapshot);
  });
});

describe('biriba sort mode persistence', () => {
  afterEach(() => localStorage.clear());

  it('defaults to original when nothing is stored', () => {
    expect(loadBiribaSortMode()).toBe('original');
  });

  it('round-trips a saved mode through localStorage', () => {
    saveBiribaSortMode('suit');
    expect(localStorage.getItem(BIRIBA_SORT_STORAGE_KEY)).toBe('suit');
    expect(loadBiribaSortMode()).toBe('suit');
  });

  it('ignores an invalid stored value', () => {
    localStorage.setItem(BIRIBA_SORT_STORAGE_KEY, 'bogus');
    expect(loadBiribaSortMode()).toBe('original');
  });
});
