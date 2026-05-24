import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { fiftyOneBestSuit, fiftyOneCardScore, fiftyOneSuitScores } from './fiftyOneSuitScores';

describe('fiftyOneCardScore', () => {
  it.each([
    [1, 11],
    [2, 2],
    [10, 10],
    [11, 10],
    [12, 10],
    [13, 10],
  ])('value %i -> %i', (v, expected) => {
    expect(fiftyOneCardScore(v)).toBe(expected);
  });
});

describe('fiftyOneSuitScores', () => {
  it('aggregates per-suit totals and skips jokers', () => {
    const cards: Card[] = [
      { design: 'SPADE', value: 1 }, // 11
      { design: 'SPADE', value: 10 }, // 10
      { design: 'HEART', value: 5 }, // 5
      { design: 'DIAMOND', value: 13 }, // 10
      { design: 'JOKER', value: 0 },
    ];
    expect(fiftyOneSuitScores(cards)).toEqual({ SPADE: 21, CLOVER: 0, HEART: 5, DIAMOND: 10 });
  });

  it('returns zeros for empty hand', () => {
    expect(fiftyOneSuitScores([])).toEqual({ SPADE: 0, CLOVER: 0, HEART: 0, DIAMOND: 0 });
  });
});

describe('fiftyOneBestSuit', () => {
  it('returns the suit with the highest total', () => {
    expect(fiftyOneBestSuit({ SPADE: 5, CLOVER: 0, HEART: 30, DIAMOND: 10 })).toBe('HEART');
  });

  it('breaks ties in S/C/H/D order', () => {
    expect(fiftyOneBestSuit({ SPADE: 10, CLOVER: 10, HEART: 10, DIAMOND: 10 })).toBe('SPADE');
  });
});
