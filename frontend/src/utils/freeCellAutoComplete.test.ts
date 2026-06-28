import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { freeCellAutoCompleteReady } from './freeCellAutoComplete';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('freeCellAutoCompleteReady', () => {
  it('is ready for an empty board', () => {
    expect(freeCellAutoCompleteReady([[], [], []])).toBe(true);
  });

  it('is ready when every column is strictly descending bottom-to-top', () => {
    const tableau = [
      [c('SPADE', 5), c('HEART', 3), c('CLOVER', 2)],
      [c('DIAMOND', 13), c('SPADE', 1)],
      [c('HEART', 7)],
    ];
    expect(freeCellAutoCompleteReady(tableau)).toBe(true);
  });

  it('is not ready when a column has a larger rank stacked on a smaller one', () => {
    const tableau = [[c('SPADE', 2), c('HEART', 5)]]; // 5 on top of 2 — blocks
    expect(freeCellAutoCompleteReady(tableau)).toBe(false);
  });

  it('is not ready on equal ranks (conservative)', () => {
    const tableau = [[c('SPADE', 4), c('HEART', 4)]];
    expect(freeCellAutoCompleteReady(tableau)).toBe(false);
  });

  it('ignores null slots when evaluating order', () => {
    const tableau = [[c('SPADE', 9), null, c('HEART', 4), null]];
    expect(freeCellAutoCompleteReady(tableau)).toBe(true);
  });
});
