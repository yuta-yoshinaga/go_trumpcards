import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { blackjackScore, blackjackSwitchPreviewScores } from './blackjackSwitchPreview';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('blackjackScore', () => {
  it('counts face cards as 10 and aces flexibly', () => {
    expect(blackjackScore([c('SPADE', 1), c('HEART', 13)])).toBe(21);
    expect(blackjackScore([c('SPADE', 1), c('HEART', 9), c('CLOVER', 5)])).toBe(15); // ace soft 1
    expect(blackjackScore([c('SPADE', 1), c('HEART', 1)])).toBe(12); // 11+11 -> 12
    expect(blackjackScore([c('SPADE', 5), c('HEART', 9), c('CLOVER', 8)])).toBe(22); // bust
  });
});

describe('blackjackSwitchPreviewScores', () => {
  it('swaps the second card between two hands', () => {
    const a: Card[] = [c('SPADE', 10), c('SPADE', 4)];
    const b: Card[] = [c('HEART', 10), c('HEART', 7)];
    // Before: a=14, b=17. After swap: a=10+7=17, b=10+4=14
    expect(blackjackSwitchPreviewScores(a, b)).toEqual({ a: 17, b: 14 });
  });

  it('returns null when either hand has fewer than 2 cards', () => {
    expect(blackjackSwitchPreviewScores([c('SPADE', 1)], [c('HEART', 2), c('HEART', 3)])).toBeNull();
  });

  it('preserves extra cards beyond the second', () => {
    const a: Card[] = [c('SPADE', 10), c('SPADE', 4), c('SPADE', 2)];
    const b: Card[] = [c('HEART', 9), c('HEART', 8)];
    // After swap: a = [S10, H8, S2] = 20; b = [H9, S4] = 13
    expect(blackjackSwitchPreviewScores(a, b)).toEqual({ a: 20, b: 13 });
  });
});
