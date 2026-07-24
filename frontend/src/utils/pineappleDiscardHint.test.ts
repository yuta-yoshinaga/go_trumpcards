import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { pineappleKeepFeatures } from './pineappleDiscardHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('pineappleKeepFeatures', () => {
  it('detects a pair (exclusive of other features)', () => {
    expect(pineappleKeepFeatures(card('SPADE', 10), card('HEART', 10))).toEqual(['pair']);
  });

  it('detects suited cards', () => {
    expect(pineappleKeepFeatures(card('SPADE', 5), card('SPADE', 9))).toEqual(['suited']);
  });

  it('detects a connector (adjacent ranks)', () => {
    expect(pineappleKeepFeatures(card('SPADE', 8), card('HEART', 9))).toEqual(['connector']);
  });

  it('detects a suited connector (both features)', () => {
    expect(pineappleKeepFeatures(card('CLOVER', 6), card('CLOVER', 7))).toEqual(['suited', 'connector']);
  });

  it('treats Ace as high (A-K is a connector)', () => {
    expect(pineappleKeepFeatures(card('SPADE', 1), card('HEART', 13))).toEqual(['connector']);
  });

  it('treats Ace as low (A-2 is a connector)', () => {
    expect(pineappleKeepFeatures(card('SPADE', 1), card('HEART', 2))).toEqual(['connector']);
  });

  it('falls back to highcard when nothing else applies', () => {
    expect(pineappleKeepFeatures(card('SPADE', 4), card('HEART', 11))).toEqual(['highcard']);
  });
});
