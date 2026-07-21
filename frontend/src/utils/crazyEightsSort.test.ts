import { afterEach, describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  CRAZYEIGHTS_SORT_STORAGE_KEY,
  type CrazyEightsSortMode,
  loadCrazyEightsSortMode,
  saveCrazyEightsSortMode,
  sortedCrazyEightsHand,
} from './crazyEightsSort';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('sortedCrazyEightsHand', () => {
  const hand = [c('DIAMOND', 5), c('SPADE', 3), c('CLOVER', 8), c('HEART', 3), c('SPADE', 13)];

  it('original mode returns identity order with matching indices', () => {
    const out = sortedCrazyEightsHand(hand, 'original');
    expect(out.map((o) => o.index)).toEqual([0, 1, 2, 3, 4]);
    expect(out.map((o) => o.card)).toEqual(hand);
  });

  it('keeps each card paired with its original index in every mode', () => {
    for (const mode of ['original', 'rank', 'suit'] as CrazyEightsSortMode[]) {
      for (const { card, index } of sortedCrazyEightsHand(hand, mode)) {
        expect(hand[index]).toBe(card);
      }
    }
  });

  it('rank mode orders by value ascending, suit as tiebreak', () => {
    const out = sortedCrazyEightsHand(hand, 'rank');
    expect(out.map((o) => o.card.value)).toEqual([3, 3, 5, 8, 13]);
    // Two 3s: SPADE (♠) sorts before HEART (♥).
    expect(out[0].card.design).toBe('SPADE');
    expect(out[1].card.design).toBe('HEART');
  });

  it('suit mode groups by suit ♠♥♦♣ with rank ascending inside each group', () => {
    const out = sortedCrazyEightsHand(hand, 'suit');
    expect(out.map((o) => o.card.design)).toEqual(['SPADE', 'SPADE', 'HEART', 'DIAMOND', 'CLOVER']);
    // Within SPADE, rank ascending: 3 before 13.
    expect(out[0].card.value).toBe(3);
    expect(out[1].card.value).toBe(13);
  });

  it('maps a display position back to the correct original index after sorting', () => {
    // The ♣8 sits at original index 2; after suit sort it renders last, but a
    // play must still target index 2 so the server removes the right card.
    const out = sortedCrazyEightsHand(hand, 'suit');
    const eight = out[out.length - 1];
    expect(eight.card).toEqual(c('CLOVER', 8));
    expect(eight.index).toBe(2);
    expect(hand[eight.index]).toBe(eight.card);
  });

  it('does not mutate the input array', () => {
    const snapshot = [...hand];
    sortedCrazyEightsHand(hand, 'rank');
    expect(hand).toEqual(snapshot);
  });
});

describe('crazy eights sort mode persistence', () => {
  afterEach(() => localStorage.clear());

  it('defaults to original when nothing is stored', () => {
    expect(loadCrazyEightsSortMode()).toBe('original');
  });

  it('round-trips a saved mode through localStorage', () => {
    saveCrazyEightsSortMode('suit');
    expect(localStorage.getItem(CRAZYEIGHTS_SORT_STORAGE_KEY)).toBe('suit');
    expect(loadCrazyEightsSortMode()).toBe('suit');
  });

  it('ignores an invalid stored value', () => {
    localStorage.setItem(CRAZYEIGHTS_SORT_STORAGE_KEY, 'bogus');
    expect(loadCrazyEightsSortMode()).toBe('original');
  });
});
