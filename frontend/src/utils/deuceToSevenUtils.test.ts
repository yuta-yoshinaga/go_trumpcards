import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { deuceToSevenBestIndices, isMadePatLow } from './deuceToSevenUtils';

const card = (design: CardDesign, value: number): Card => ({ design, value });

describe('isMadePatLow', () => {
  it('returns true for the nut low 7-5-4-3-2', () => {
    expect(
      isMadePatLow([card('SPADE', 7), card('HEART', 5), card('DIAMOND', 4), card('CLOVER', 3), card('SPADE', 2)]),
    ).toBe(true);
  });

  it('returns true for an 8-or-better low', () => {
    expect(
      isMadePatLow([card('SPADE', 8), card('HEART', 6), card('DIAMOND', 4), card('CLOVER', 3), card('SPADE', 2)]),
    ).toBe(true);
  });

  it('returns false when the high card exceeds 8', () => {
    expect(
      isMadePatLow([card('SPADE', 9), card('HEART', 6), card('DIAMOND', 4), card('CLOVER', 3), card('SPADE', 2)]),
    ).toBe(false);
  });

  it('returns false for a pair', () => {
    expect(
      isMadePatLow([card('SPADE', 2), card('HEART', 2), card('DIAMOND', 4), card('CLOVER', 6), card('SPADE', 8)]),
    ).toBe(false);
  });

  it('returns false for a flush', () => {
    expect(
      isMadePatLow([card('SPADE', 2), card('SPADE', 4), card('SPADE', 5), card('SPADE', 6), card('SPADE', 8)]),
    ).toBe(false);
  });

  it('returns false for a straight', () => {
    expect(
      isMadePatLow([card('SPADE', 2), card('HEART', 3), card('DIAMOND', 4), card('CLOVER', 5), card('SPADE', 6)]),
    ).toBe(false);
  });

  it('treats the ace as high, so A-2-3-4-5 is not a made low', () => {
    expect(
      isMadePatLow([card('SPADE', 1), card('HEART', 2), card('DIAMOND', 3), card('CLOVER', 4), card('SPADE', 5)]),
    ).toBe(false);
  });

  it('returns false when fewer than 5 cards are supplied', () => {
    expect(isMadePatLow([card('SPADE', 7), card('HEART', 5)])).toBe(false);
  });

  it('returns false when a Joker appears', () => {
    expect(
      isMadePatLow([card('JOKER', 0), card('HEART', 5), card('DIAMOND', 4), card('CLOVER', 3), card('SPADE', 2)]),
    ).toBe(false);
  });
});

describe('deuceToSevenBestIndices', () => {
  it('keeps every index for a made pat low', () => {
    const hand = [card('SPADE', 7), card('HEART', 5), card('DIAMOND', 4), card('CLOVER', 3), card('SPADE', 2)];
    expect(deuceToSevenBestIndices(hand)).toEqual([0, 1, 2, 3, 4]);
  });

  it('drops high cards (9+) and the Ace as draw candidates', () => {
    // 2 (idx0), 4 (idx2), 5 (idx3) are low keepers; K (idx1) and A (idx4) are high → dropped.
    const hand = [card('SPADE', 2), card('HEART', 13), card('DIAMOND', 4), card('CLOVER', 5), card('SPADE', 1)];
    expect(deuceToSevenBestIndices(hand)).toEqual([0, 2, 3]);
  });

  it('keeps only one card of a low pair', () => {
    // Two 4s (idx1, idx3): keep exactly one; 2 and 8 also kept.
    const hand = [card('SPADE', 2), card('HEART', 4), card('DIAMOND', 8), card('CLOVER', 4), card('SPADE', 10)];
    const out = deuceToSevenBestIndices(hand);
    expect(out).toContain(0); // 2
    expect(out).toContain(2); // 8
    expect(out).not.toContain(4); // 10 is high → dropped
    expect([out.includes(1), out.includes(3)].filter(Boolean)).toHaveLength(1);
  });

  it('returns an empty array for an empty hand', () => {
    expect(deuceToSevenBestIndices([])).toEqual([]);
  });
});
