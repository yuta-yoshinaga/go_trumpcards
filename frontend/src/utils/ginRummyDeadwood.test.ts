import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  bestDeadwoodValue,
  bestMeldSplit,
  calcDeadwoodValue,
  ginRummyCardValue,
  ginRummyMeldLabel,
  ginRummyScoreBreakdown,
} from './ginRummyDeadwood';

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

describe('bestMeldSplit', () => {
  const sorted = (s: ReadonlySet<number>) => [...s].sort((a, b) => a - b);

  it('returns an empty split for an empty hand', () => {
    const split = bestMeldSplit([]);
    expect(sorted(split.meldedIndices)).toEqual([]);
    expect(split.deadwoodValue).toBe(0);
  });

  it('melds no card when no meld exists', () => {
    const split = bestMeldSplit([c('SPADE', 2), c('HEART', 5), c('CLOVER', 9)]);
    expect(sorted(split.meldedIndices)).toEqual([]);
    expect(split.deadwoodValue).toBe(2 + 5 + 9);
  });

  it('marks a set of 3 as melded, leaving the stray as deadwood', () => {
    // indices 0,1,2 = three 7s (set); index 3 = stray 4 (deadwood)
    const split = bestMeldSplit([c('SPADE', 7), c('HEART', 7), c('CLOVER', 7), c('DIAMOND', 4)]);
    expect(sorted(split.meldedIndices)).toEqual([0, 1, 2]);
    expect(split.deadwoodValue).toBe(4);
  });

  it('marks a run of 3 as melded, leaving the stray as deadwood', () => {
    // indices 0,1,2 = ♠5-6-7 run; index 3 = stray K (deadwood)
    const split = bestMeldSplit([c('SPADE', 5), c('SPADE', 6), c('SPADE', 7), c('HEART', 13)]);
    expect(sorted(split.meldedIndices)).toEqual([0, 1, 2]);
    expect(split.deadwoodValue).toBe(10);
  });

  it('prefers the run over an impossible set with duplicate ranks', () => {
    // ♠7-8-9 run (0,1,2); leftover ♥7,♣7 (3,4) are only 2 sevens → no set → deadwood
    const split = bestMeldSplit([c('SPADE', 7), c('SPADE', 8), c('SPADE', 9), c('HEART', 7), c('CLOVER', 7)]);
    expect(sorted(split.meldedIndices)).toEqual([0, 1, 2]);
    expect(split.deadwoodValue).toBe(14);
  });

  it('melds every card when the hand is fully melded', () => {
    const hand: Card[] = [
      c('SPADE', 5),
      c('SPADE', 6),
      c('SPADE', 7),
      c('SPADE', 8),
      c('HEART', 9),
      c('CLOVER', 9),
      c('DIAMOND', 9),
    ];
    const split = bestMeldSplit(hand);
    expect(sorted(split.meldedIndices)).toEqual([0, 1, 2, 3, 4, 5, 6]);
    expect(split.deadwoodValue).toBe(0);
  });

  it('stays consistent with bestDeadwoodValue', () => {
    const hand = [c('SPADE', 7), c('SPADE', 8), c('SPADE', 9), c('HEART', 7), c('CLOVER', 7)];
    expect(bestMeldSplit(hand).deadwoodValue).toBe(bestDeadwoodValue(hand));
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

describe('ginRummyScoreBreakdown', () => {
  const knocker = { id: 0, cards: [] as Card[] };
  const opp = (cards: Card[]) => ({ id: 1, cards });

  it('scores a plain knock as the deadwood difference to the knocker', () => {
    // knocker deadwood 9, opponent 5+6+8=19 (no meld) → 19-9=10, no bonus.
    const b = ginRummyScoreBreakdown(
      [knocker, opp([c('DIAMOND', 5), c('CLOVER', 6), c('HEART', 8)])],
      0,
      [c('CLOVER', 9)],
      false,
    );
    expect(b).toEqual({
      outcome: 'knock',
      winnerId: 0,
      knockerDeadwood: 9,
      opponentDeadwood: 19,
      base: 10,
      bonus: 0,
      total: 10,
    });
  });

  it('scores a gin as opponent deadwood + 25 bonus to the knocker', () => {
    const b = ginRummyScoreBreakdown([knocker, opp([c('DIAMOND', 5), c('CLOVER', 6), c('HEART', 8)])], 0, [], true);
    expect(b).toMatchObject({ outcome: 'gin', winnerId: 0, opponentDeadwood: 19, base: 19, bonus: 25, total: 44 });
  });

  it('scores an undercut as the difference + 25 bonus to the defender', () => {
    // knocker deadwood 9, opponent ♦5-6-7 run + ♥2 = 2 ≤ 9 → 7 diff + 25 = 32 to opponent.
    const b = ginRummyScoreBreakdown(
      [knocker, opp([c('DIAMOND', 5), c('DIAMOND', 6), c('DIAMOND', 7), c('HEART', 2)])],
      0,
      [c('CLOVER', 9)],
      false,
    );
    expect(b).toEqual({
      outcome: 'undercut',
      winnerId: 1,
      knockerDeadwood: 9,
      opponentDeadwood: 2,
      base: 7,
      bonus: 25,
      total: 32,
    });
  });

  it('treats equal deadwood as an undercut for the defender', () => {
    const b = ginRummyScoreBreakdown([knocker, opp([c('CLOVER', 9)])], 0, [c('SPADE', 9)], false);
    expect(b).toMatchObject({ outcome: 'undercut', winnerId: 1, base: 0, bonus: 25, total: 25 });
  });

  it('returns null for a drawn round (no knocker)', () => {
    expect(ginRummyScoreBreakdown([knocker, opp([c('CLOVER', 9)])], -1, [], false)).toBeNull();
  });

  it('returns null when the players cannot be resolved', () => {
    expect(ginRummyScoreBreakdown([knocker], 0, [], false)).toBeNull();
  });
});
