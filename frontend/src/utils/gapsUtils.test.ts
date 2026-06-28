import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { gapsLockedPrefixLengths, gapsLockedTotal } from './gapsUtils';

const card = (design: CardDesign, value: number): Card => ({ design, value });

describe('gapsLockedPrefixLengths', () => {
  it('returns 0 for a row with no leading 2', () => {
    expect(gapsLockedPrefixLengths([[card('SPADE', 5), card('SPADE', 6)]])).toEqual([0]);
  });

  it('returns 0 for a row whose first cell is empty', () => {
    expect(gapsLockedPrefixLengths([[null, card('SPADE', 2)]])).toEqual([0]);
  });

  it('counts the run while suit + sequential value hold', () => {
    const row = [card('HEART', 2), card('HEART', 3), card('HEART', 4), card('SPADE', 5), card('HEART', 6)];
    expect(gapsLockedPrefixLengths([row])).toEqual([3]);
  });

  it('stops the lock at the first suit mismatch even if the value sequence is correct', () => {
    // The 3 is sequential (2 + 1) but the suit changes — must break.
    const row = [card('SPADE', 2), card('HEART', 3), card('HEART', 4)];
    expect(gapsLockedPrefixLengths([row])).toEqual([1]);
  });

  it('stops the lock at the first gap', () => {
    const row = [card('HEART', 2), card('HEART', 3), null, card('HEART', 4)];
    expect(gapsLockedPrefixLengths([row])).toEqual([2]);
  });

  it('handles a full 13-card run', () => {
    const row: (Card | null)[] = [];
    for (let v = 2; v <= 13; v++) row.push(card('SPADE', v));
    expect(gapsLockedPrefixLengths([row])).toEqual([row.length]);
  });

  it('computes a value per row independently', () => {
    const grid = [[card('SPADE', 2), card('SPADE', 3)], [card('HEART', 5)], [], [card('CLOVER', 2)]];
    expect(gapsLockedPrefixLengths(grid)).toEqual([2, 0, 0, 1]);
  });
});

describe('gapsLockedTotal', () => {
  it('sums the locked prefix lengths across all rows', () => {
    const grid = [[card('SPADE', 2), card('SPADE', 3)], [card('HEART', 5)], [], [card('CLOVER', 2)]];
    expect(gapsLockedTotal(grid)).toBe(3);
  });

  it('returns 0 when no row starts with a 2', () => {
    expect(gapsLockedTotal([[card('HEART', 5)], [card('SPADE', 7)]])).toBe(0);
  });
});
