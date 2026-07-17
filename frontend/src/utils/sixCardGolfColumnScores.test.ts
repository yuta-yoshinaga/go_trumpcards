import { describe, expect, it } from 'vitest';
import type { Card, SixCardGolfSlot } from '../types/card';
import { sixCardGolfCardScore, sixCardGolfColumnScores } from './sixCardGolfColumnScores';

const card = (value: number): Card => ({ design: 'SPADE', value }) as unknown as Card;
const slot = (value: number | null, faceUp = true): SixCardGolfSlot => ({
  card: value === null ? null : card(value),
  faceUp,
});

describe('sixCardGolfCardScore', () => {
  it('scores K as 0, A as 1, J/Q as 10, and others at face value', () => {
    expect(sixCardGolfCardScore(card(13))).toBe(0);
    expect(sixCardGolfCardScore(card(1))).toBe(1);
    expect(sixCardGolfCardScore(card(11))).toBe(10);
    expect(sixCardGolfCardScore(card(12))).toBe(10);
    expect(sixCardGolfCardScore(card(7))).toBe(7);
    expect(sixCardGolfCardScore(null)).toBe(0);
  });
});

describe('sixCardGolfColumnScores', () => {
  it('returns one score per column', () => {
    const grid = [slot(3), slot(5), slot(7), slot(4), slot(6), slot(8)];
    const cols = sixCardGolfColumnScores(grid);
    expect(cols).toHaveLength(3);
    expect(cols.map((c) => c.score)).toEqual([7, 11, 15]);
    expect(cols.every((c) => !c.isPair)).toBe(true);
    // All cards face up → no column is uncertain.
    expect(cols.every((c) => !c.hasHidden)).toBe(true);
  });

  it('cancels a matched column to zero', () => {
    // Column 0 is a pair (5 over 5) → 0; column 1 is K over K → pair 0.
    const grid = [slot(5), slot(13), slot(2), slot(5), slot(13), slot(9)];
    const cols = sixCardGolfColumnScores(grid);
    expect(cols[0]).toEqual({ score: 0, isPair: true, hasHidden: false });
    expect(cols[1]).toEqual({ score: 0, isPair: true, hasHidden: false });
    expect(cols[2]).toEqual({ score: 11, isPair: false, hasHidden: false });
  });

  it('ignores face-down cards in the total and flags the column uncertain', () => {
    const grid = [slot(8), slot(4), slot(2), slot(6, false), slot(4, false), slot(2)];
    const cols = sixCardGolfColumnScores(grid);
    // col0: 8 + (face-down) = 8, uncertain; col1: 4 + (face-down) = 4, uncertain; col2: 2+2 pair → 0, certain.
    expect(cols[0]).toEqual({ score: 8, isPair: false, hasHidden: true });
    expect(cols[1]).toEqual({ score: 4, isPair: false, hasHidden: true });
    expect(cols[2]).toEqual({ score: 0, isPair: true, hasHidden: false });
  });

  it('marks a column uncertain when a slot has no card yet', () => {
    const grid = [slot(8), slot(4), slot(2), slot(null), slot(6), slot(9)];
    const cols = sixCardGolfColumnScores(grid);
    expect(cols[0].hasHidden).toBe(true); // bottom slot has no card
    expect(cols[1].hasHidden).toBe(false);
    expect(cols[2].hasHidden).toBe(false);
  });
});
