import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { rummy500CardPenalty, rummy500HandPenalty } from './rummy500HandPenalty';

describe('rummy500CardPenalty', () => {
  it.each([
    [1, 1],
    [5, 5],
    [9, 9],
    [10, 10],
    [11, 10],
    [13, 10],
  ])('value %i -> %i', (v, expected) => {
    expect(rummy500CardPenalty(v)).toBe(expected);
  });
});

describe('rummy500HandPenalty', () => {
  it('sums per-card penalty across the hand', () => {
    const cards: Card[] = [
      { design: 'SPADE', value: 13 },
      { design: 'HEART', value: 10 },
      { design: 'CLOVER', value: 1 },
      { design: 'DIAMOND', value: 4 },
    ];
    expect(rummy500HandPenalty(cards)).toBe(10 + 10 + 1 + 4);
  });
  it('returns 0 for an empty hand', () => {
    expect(rummy500HandPenalty([])).toBe(0);
  });
});
