import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { karnoffelRankKey, suitIndex } from './karnoffelRanks';

const c = (design: CardDesign, value: number): Card => ({ design, value });

describe('suitIndex', () => {
  it('matches the backend numbering used by chosenSuit', () => {
    expect([suitIndex('SPADE'), suitIndex('CLOVER'), suitIndex('HEART'), suitIndex('DIAMOND')]).toEqual([1, 2, 3, 4]);
  });

  it('is -1 for a design with no suit', () => {
    expect(suitIndex('JOKER')).toBe(-1);
  });
});

describe('karnoffelRankKey', () => {
  it('names every titled rank of the chosen suit', () => {
    const titles = [
      [11, 'karnoffel'],
      [7, 'devil'],
      [6, 'pope'],
      [2, 'kaiser'],
      [3, 'oberstecher'],
      [4, 'unterstecher'],
      [5, 'farbenstecher'],
    ] as const;
    for (const [value, key] of titles) {
      expect(karnoffelRankKey(c('HEART', value), 3)).toBe(key);
    }
  });

  it('gives the same rank no title outside the chosen suit', () => {
    // The whole point: a Jack is only the Karnöffel in this deal's suit.
    expect(karnoffelRankKey(c('SPADE', 11), 3)).toBeNull();
  });

  it('gives untitled ranks of the chosen suit no title', () => {
    for (const value of [8, 9, 10, 12, 13, 1]) {
      expect(karnoffelRankKey(c('HEART', value), 3)).toBeNull();
    }
  });
});
