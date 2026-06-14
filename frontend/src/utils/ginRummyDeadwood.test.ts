import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { bestDeadwoodValue, calcDeadwoodValue, ginRummyCardValue, ginRummyMeldLabel } from './ginRummyDeadwood';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('ginRummyCardValue', () => {
  it('A=1, 2-9=face, 10/J/Q/K=10', () => {
    expect(ginRummyCardValue(c('SPADE', 1))).toBe(1);
    expect(ginRummyCardValue(c('SPADE', 7))).toBe(7);
    expect(ginRummyCardValue(c('SPADE', 10))).toBe(10);
    expect(ginRummyCardValue(c('SPADE', 11))).toBe(10);
    expect(ginRummyCardValue(c('SPADE', 13))).toBe(10);
  });
});

describe('calcDeadwoodValue', () => {
  it('sums Gin Rummy card values', () => {
    expect(calcDeadwoodValue([c('SPADE', 7), c('HEART', 11), c('CLOVER', 1)])).toBe(7 + 10 + 1);
  });
});

describe('bestDeadwoodValue', () => {
  it('returns 0 on empty hand', () => {
    expect(bestDeadwoodValue([])).toBe(0);
  });

  it('counts entire hand as deadwood when no meld exists', () => {
    expect(bestDeadwoodValue([c('SPADE', 2), c('HEART', 5), c('CLOVER', 9)])).toBe(2 + 5 + 9);
  });

  it('removes a set of 3 from deadwood', () => {
    // 3 sevens (a set) + a stray 4 → only the 4 (value 4) is deadwood
    expect(bestDeadwoodValue([c('SPADE', 7), c('HEART', 7), c('CLOVER', 7), c('DIAMOND', 4)])).toBe(4);
  });

  it('removes a run of 3 from deadwood', () => {
    // ♠5-6-7 run + stray K → only K (10) deadwood
    expect(bestDeadwoodValue([c('SPADE', 5), c('SPADE', 6), c('SPADE', 7), c('HEART', 13)])).toBe(10);
  });

  it('picks the meld split that minimises deadwood', () => {
    // ♠7-8-9 (run) + ♥7-♣7 leftover (only 2 sevens → no set).
    // Best choice = run; remaining 7+7 = 14 deadwood.
    expect(bestDeadwoodValue([c('SPADE', 7), c('SPADE', 8), c('SPADE', 9), c('HEART', 7), c('CLOVER', 7)])).toBe(14);
  });

  it('reports 0 when the hand is fully melded', () => {
    // ♠5-6-7-8 run + 9♥-9♣-9♦ set → 7 cards, fully melded
    const hand: Card[] = [
      c('SPADE', 5),
      c('SPADE', 6),
      c('SPADE', 7),
      c('SPADE', 8),
      c('HEART', 9),
      c('CLOVER', 9),
      c('DIAMOND', 9),
    ];
    expect(bestDeadwoodValue(hand)).toBe(0);
  });
});

describe('ginRummyMeldLabel', () => {
  it('labels a same-rank set with the rank name', () => {
    expect(ginRummyMeldLabel([c('SPADE', 3), c('HEART', 3), c('DIAMOND', 3)])).toBe('3');
    expect(ginRummyMeldLabel([c('SPADE', 13), c('HEART', 13), c('CLOVER', 13)])).toBe('K');
    expect(ginRummyMeldLabel([c('SPADE', 1), c('HEART', 1), c('CLOVER', 1)])).toBe('A');
  });

  it('labels a same-suit run with the suit symbol and low-high range', () => {
    expect(ginRummyMeldLabel([c('SPADE', 5), c('SPADE', 3), c('SPADE', 4)])).toBe('♠ 3-5');
    expect(ginRummyMeldLabel([c('HEART', 1), c('HEART', 2), c('HEART', 3)])).toBe('♥ A-3');
  });

  it('returns an empty string for an empty meld', () => {
    expect(ginRummyMeldLabel([])).toBe('');
  });
});
