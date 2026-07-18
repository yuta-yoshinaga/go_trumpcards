import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  bestChinchonDeadwoodValue,
  bestChinchonMeldSplit,
  calcChinchonDeadwoodValue,
  chinchonCardValue,
  chinchonDeadwoodBreakdown,
  chinchonMeldLabel,
  chinchonRankPosition,
} from './chinchonDeadwood';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('chinchonRankPosition', () => {
  it('maps A-7 to 1-7 and J/Q/K to 8/9/10 (7 and J adjacent)', () => {
    expect(chinchonRankPosition(1)).toBe(1);
    expect(chinchonRankPosition(7)).toBe(7);
    expect(chinchonRankPosition(11)).toBe(8);
    expect(chinchonRankPosition(12)).toBe(9);
    expect(chinchonRankPosition(13)).toBe(10);
  });

  it('returns 0 for ranks absent from the 40-card deck', () => {
    expect(chinchonRankPosition(8)).toBe(0);
    expect(chinchonRankPosition(10)).toBe(0);
  });
});

describe('chinchonCardValue', () => {
  it('scores A=1, face value 2-7, J/Q/K=10', () => {
    expect(chinchonCardValue(c('SPADE', 1))).toBe(1);
    expect(chinchonCardValue(c('SPADE', 5))).toBe(5);
    expect(chinchonCardValue(c('SPADE', 11))).toBe(10);
    expect(chinchonCardValue(c('SPADE', 13))).toBe(10);
  });
});

describe('calcChinchonDeadwoodValue', () => {
  it('sums card values', () => {
    expect(calcChinchonDeadwoodValue([c('SPADE', 11), c('HEART', 1)])).toBe(11);
  });
});

describe('chinchonMeldLabel', () => {
  it('returns empty string for empty meld', () => {
    expect(chinchonMeldLabel([])).toBe('');
  });

  it('labels a same-rank set with the rank', () => {
    expect(chinchonMeldLabel([c('SPADE', 5), c('HEART', 5), c('CLOVER', 5)])).toBe('5');
  });

  it('labels a 7-J-Q run across the 7/J gap', () => {
    const label = chinchonMeldLabel([c('SPADE', 7), c('SPADE', 11), c('SPADE', 12)]);
    expect(label).toContain('7');
    expect(label).toContain('Q');
  });
});

describe('bestChinchonDeadwoodValue', () => {
  it('returns 0 for an empty hand', () => {
    expect(bestChinchonDeadwoodValue([])).toBe(0);
  });

  it('returns full value when no meld exists', () => {
    expect(bestChinchonDeadwoodValue([c('SPADE', 11), c('HEART', 1)])).toBe(11);
  });

  it('removes a same-suit run spanning the 7/J gap from the deadwood', () => {
    // ♠7-J-Q run (positions 7,8,9) + lone ♥A → deadwood = 1
    const hand = [c('SPADE', 7), c('SPADE', 11), c('SPADE', 12), c('HEART', 1)];
    expect(bestChinchonDeadwoodValue(hand)).toBe(1);
  });

  it('removes a same-rank set from the deadwood', () => {
    const hand = [c('SPADE', 5), c('HEART', 5), c('CLOVER', 5), c('DIAMOND', 11)];
    expect(bestChinchonDeadwoodValue(hand)).toBe(10);
  });
});

describe('chinchonDeadwoodBreakdown', () => {
  it('lists the leftover deadwood cards and their values after the best meld split', () => {
    // ♠5-♥5-♣5 set melds away; ♦K(10) and ♥3(3) are deadwood.
    const hand = [c('SPADE', 5), c('HEART', 5), c('CLOVER', 5), c('DIAMOND', 13), c('HEART', 3)];
    const bd = chinchonDeadwoodBreakdown(hand);
    expect(bd.total).toBe(13);
    expect([...bd.values].sort((a, b) => a - b)).toEqual([3, 10]);
    expect(bd.cards).toHaveLength(2);
  });

  it('returns an empty breakdown when every card melds', () => {
    const hand = [c('SPADE', 5), c('HEART', 5), c('CLOVER', 5)];
    expect(chinchonDeadwoodBreakdown(hand)).toEqual({ cards: [], values: [], total: 0 });
  });
});

describe('bestChinchonMeldSplit', () => {
  it('marks the melded card indices and leaves deadwood out', () => {
    // ♠5-♥5-♣5 set (idx 0,1,2) melds; ♦K(idx 3) and ♥3(idx 4) are deadwood.
    const hand = [c('SPADE', 5), c('HEART', 5), c('CLOVER', 5), c('DIAMOND', 13), c('HEART', 3)];
    const split = bestChinchonMeldSplit(hand);
    expect([...split.meldedIndices].sort((a, b) => a - b)).toEqual([0, 1, 2]);
    expect(split.meldedIndices.has(3)).toBe(false);
    expect(split.meldedIndices.has(4)).toBe(false);
    expect(split.deadwoodValue).toBe(13);
  });

  it('marks a same-suit run spanning the 7/J gap as melded', () => {
    // ♠7-J-Q run (idx 0,1,2) melds; ♥A(idx 3) is deadwood.
    const hand = [c('SPADE', 7), c('SPADE', 11), c('SPADE', 12), c('HEART', 1)];
    const split = bestChinchonMeldSplit(hand);
    expect([...split.meldedIndices].sort((a, b) => a - b)).toEqual([0, 1, 2]);
    expect(split.deadwoodValue).toBe(1);
  });

  it('returns an empty set for an empty hand', () => {
    const split = bestChinchonMeldSplit([]);
    expect(split.meldedIndices.size).toBe(0);
    expect(split.deadwoodValue).toBe(0);
  });

  it('marks every index as deadwood when no meld exists', () => {
    const hand = [c('SPADE', 5), c('HEART', 13)];
    const split = bestChinchonMeldSplit(hand);
    expect(split.meldedIndices.size).toBe(0);
    expect(split.deadwoodValue).toBe(15);
  });
});
