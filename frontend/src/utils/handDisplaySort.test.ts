import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { type HandSortMode, sortedHandForDisplay } from './handDisplaySort';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('sortedHandForDisplay', () => {
  const hand = [c('DIAMOND', 5), c('SPADE', 3), c('JOKER', 0), c('HEART', 3), c('SPADE', 13)];

  it('original mode returns identity order with matching indices', () => {
    const out = sortedHandForDisplay(hand, 'original');
    expect(out.map((o) => o.index)).toEqual([0, 1, 2, 3, 4]);
    expect(out.map((o) => o.card)).toEqual(hand);
  });

  it('keeps each card paired with its original index in every mode', () => {
    for (const mode of ['original', 'rank', 'suit'] as HandSortMode[]) {
      for (const { card, index } of sortedHandForDisplay(hand, mode)) {
        expect(hand[index]).toBe(card);
      }
    }
  });

  it('rank mode orders by value ascending with the joker last', () => {
    const out = sortedHandForDisplay(hand, 'rank');
    expect(out.map((o) => o.card.value)).toEqual([3, 3, 5, 13, 0]);
    expect(out[out.length - 1].card.design).toBe('JOKER');
  });

  it('suit mode groups by suit ♠♥♦♣ with the joker last', () => {
    const out = sortedHandForDisplay(hand, 'suit');
    expect(out.map((o) => o.card.design)).toEqual(['SPADE', 'SPADE', 'HEART', 'DIAMOND', 'JOKER']);
    // Within SPADE, rank ascending: 3 before 13.
    expect(out[0].card.value).toBe(3);
    expect(out[1].card.value).toBe(13);
  });

  it('does not mutate the input array', () => {
    const snapshot = [...hand];
    sortedHandForDisplay(hand, 'rank');
    expect(hand).toEqual(snapshot);
  });
});
