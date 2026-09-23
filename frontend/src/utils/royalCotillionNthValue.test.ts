import { describe, expect, it } from 'vitest';
import { CARD_VALUE_MAX, royalCotillionNextRank, royalCotillionNthValue } from './royalCotillionNthValue';

describe('royalCotillionNthValue', () => {
  it('generates the exact sequence for Ace-started foundations', () => {
    const expected = [1, 3, 5, 7, 9, 11, 13, 2, 4, 6, 8, 10, 12];
    const actual = Array.from({ length: 13 }, (_, n) => royalCotillionNthValue(1, n));
    expect(actual).toEqual(expected);
  });

  it('generates the exact sequence for Two-started foundations', () => {
    const expected = [2, 4, 6, 8, 10, 12, 1, 3, 5, 7, 9, 11, 13];
    const actual = Array.from({ length: 13 }, (_, n) => royalCotillionNthValue(2, n));
    expect(actual).toEqual(expected);
  });

  it('wraps around correctly at boundary: K (13) is followed by 2 for Ace-start', () => {
    // n=6 is K (13), n=7 wraps to 2
    expect(royalCotillionNthValue(1, 6)).toBe(13);
    expect(royalCotillionNthValue(1, 7)).toBe(2);
  });

  it('wraps around correctly at boundary: Q (12) is followed by A (1) for Two-start', () => {
    // n=5 is Q (12), n=6 wraps to A (1)
    expect(royalCotillionNthValue(2, 5)).toBe(12);
    expect(royalCotillionNthValue(2, 6)).toBe(1);
  });

  it('returns to the starting value after a full cycle of 13 cards', () => {
    // Cycle for Ace-start
    expect(royalCotillionNthValue(1, 0)).toBe(1);
    expect(royalCotillionNthValue(1, 13)).toBe(1);
    expect(royalCotillionNthValue(1, 26)).toBe(1);

    // Cycle for Two-start
    expect(royalCotillionNthValue(2, 0)).toBe(2);
    expect(royalCotillionNthValue(2, 13)).toBe(2);
    expect(royalCotillionNthValue(2, 26)).toBe(2);
  });

  it('covers all 13 distinct card values without duplicates in 13 steps', () => {
    const aceSet = new Set(Array.from({ length: 13 }, (_, n) => royalCotillionNthValue(1, n)));
    expect(aceSet.size).toBe(CARD_VALUE_MAX);

    const twoSet = new Set(Array.from({ length: 13 }, (_, n) => royalCotillionNthValue(2, n)));
    expect(twoSet.size).toBe(CARD_VALUE_MAX);
  });
});

describe('royalCotillionNextRank', () => {
  it('returns starting rank when foundation is empty', () => {
    expect(royalCotillionNextRank(0, true)).toBe(1); // Ace-start
    expect(royalCotillionNextRank(0, false)).toBe(2); // Two-start
  });

  it('returns the next rank for partially filled foundations', () => {
    expect(royalCotillionNextRank(1, true)).toBe(3);
    expect(royalCotillionNextRank(1, false)).toBe(4);
  });

  it('returns correct wrapped ranks after boundary cards are placed', () => {
    // Ace-start: 7 cards placed (A, 3, 5, 7, 9, J, K), next is 2
    expect(royalCotillionNextRank(7, true)).toBe(2);

    // Two-start: 6 cards placed (2, 4, 6, 8, 10, Q), next is 1 (A)
    expect(royalCotillionNextRank(6, false)).toBe(1);
  });

  it('returns null when foundation is complete (13 cards)', () => {
    expect(royalCotillionNextRank(13, true)).toBeNull();
    expect(royalCotillionNextRank(13, false)).toBeNull();
    expect(royalCotillionNextRank(14, true)).toBeNull();
  });
});
