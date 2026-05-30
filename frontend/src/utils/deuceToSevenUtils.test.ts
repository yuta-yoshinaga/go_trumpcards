import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { isMadePatLow } from './deuceToSevenUtils';

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
