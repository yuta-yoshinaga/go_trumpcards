import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { stalactitesAutoCompleteReady } from './stalactitesAutoComplete';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('stalactitesAutoCompleteReady', () => {
  it('is ready for an empty board', () => {
    expect(stalactitesAutoCompleteReady([[], [], []], 1)).toBe(true);
  });

  it('is ready when every column descends in foundation order bottom-to-top', () => {
    const tableau = [
      [c('SPADE', 5), c('HEART', 3), c('CLOVER', 2)],
      [c('DIAMOND', 13), c('SPADE', 1)],
      [c('HEART', 7)],
    ];
    expect(stalactitesAutoCompleteReady(tableau, 1)).toBe(true);
  });

  it('is not ready when a column has a later card stacked on an earlier one', () => {
    const tableau = [[c('SPADE', 2), c('HEART', 5)]]; // 5 on top of 2 — blocks
    expect(stalactitesAutoCompleteReady(tableau, 1)).toBe(false);
  });

  it('is not ready on equal ranks (conservative)', () => {
    const tableau = [[c('SPADE', 4), c('HEART', 4)]];
    expect(stalactitesAutoCompleteReady(tableau, 4)).toBe(false);
  });

  it('ignores null slots when evaluating order', () => {
    const tableau = [[c('SPADE', 9), null, c('HEART', 4), null]];
    expect(stalactitesAutoCompleteReady(tableau, 1)).toBe(true);
  });

  // **Foundation order wraps, so raw rank is the wrong comparison.** The King
  // over Ace column below is "descending" by rank and was reported ready by the
  // FreeCell-derived version. With base rank 7 the sequence is 7,8,...,K,A,...,
  // so the King must be played BEFORE the Ace and is buried: auto-complete
  // stalls, and promising a mop-up here would be a lie.
  it('is not ready when the wrap makes a rank-descending column unplayable', () => {
    const tableau = [[c('DIAMOND', 13), c('SPADE', 1)]];
    expect(stalactitesAutoCompleteReady(tableau, 7)).toBe(false);
    // The same column IS ready when the base rank is Ace, where rank order and
    // foundation order coincide.
    expect(stalactitesAutoCompleteReady(tableau, 1)).toBe(true);
  });

  it('is ready for a column that only descends once the wrap is applied', () => {
    // Base 7: order(7)=0, order(6)=12. A 6 under a 7 is descending in
    // foundation order even though 6 < 7 by rank.
    const tableau = [[c('SPADE', 6), c('HEART', 7)]];
    expect(stalactitesAutoCompleteReady(tableau, 7)).toBe(true);
    expect(stalactitesAutoCompleteReady(tableau, 1)).toBe(false);
  });
});
