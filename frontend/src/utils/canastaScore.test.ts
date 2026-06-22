import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { canastaCardValue, canastaMinMeld, canastaSelectionPoints } from './canastaScore';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('canastaMinMeld', () => {
  it('returns the threshold for each cumulative-score band', () => {
    expect(canastaMinMeld(-50)).toBe(15);
    expect(canastaMinMeld(0)).toBe(50);
    expect(canastaMinMeld(1499)).toBe(50);
    expect(canastaMinMeld(1500)).toBe(90);
    expect(canastaMinMeld(2999)).toBe(90);
    expect(canastaMinMeld(3000)).toBe(120);
  });
});

describe('canastaCardValue', () => {
  it('scores jokers, wild 2s, aces, black 3s, high and low cards', () => {
    expect(canastaCardValue(c('JOKER', 0))).toBe(50);
    expect(canastaCardValue(c('SPADE', 2))).toBe(20);
    expect(canastaCardValue(c('HEART', 1))).toBe(20);
    expect(canastaCardValue(c('SPADE', 3))).toBe(5); // black 3
    expect(canastaCardValue(c('HEART', 8))).toBe(10);
    expect(canastaCardValue(c('CLOVER', 13))).toBe(10);
    expect(canastaCardValue(c('DIAMOND', 5))).toBe(5);
  });
});

describe('canastaSelectionPoints', () => {
  it('sums the card values', () => {
    expect(canastaSelectionPoints([c('JOKER', 0), c('SPADE', 1), c('HEART', 8)])).toBe(80);
    expect(canastaSelectionPoints([])).toBe(0);
  });
});
