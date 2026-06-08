import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { classifyTienLenCombo, isValidTienLenCombo } from './tienLenComboValidator';

const c = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

describe('classifyTienLenCombo', () => {
  it('classifies singles, pairs, triples and four-of-a-kind', () => {
    expect(classifyTienLenCombo([c('SPADE', 7)])).toBe('single');
    expect(classifyTienLenCombo([c('SPADE', 7), c('HEART', 7)])).toBe('pair');
    expect(classifyTienLenCombo([c('SPADE', 7), c('HEART', 7), c('CLOVER', 7)])).toBe('triple');
    expect(classifyTienLenCombo([c('SPADE', 7), c('HEART', 7), c('CLOVER', 7), c('DIAMOND', 7)])).toBe('fourOfAKind');
  });

  it('classifies straights of length >= 3 (mixed suits), rejecting any containing a 2', () => {
    expect(classifyTienLenCombo([c('SPADE', 3), c('HEART', 4), c('CLOVER', 5)])).toBe('straight');
    // A (1) is high but legal in a run J-Q-K-A; 2 is not.
    expect(classifyTienLenCombo([c('SPADE', 11), c('HEART', 12), c('CLOVER', 13), c('DIAMOND', 1)])).toBe('straight');
    expect(classifyTienLenCombo([c('SPADE', 13), c('HEART', 1), c('CLOVER', 2)])).toBe('invalid'); // contains a 2
  });

  it('classifies a three-pair run (chop) and rejects non-consecutive pairs', () => {
    expect(
      classifyTienLenCombo([c('SPADE', 4), c('HEART', 4), c('SPADE', 5), c('HEART', 5), c('SPADE', 6), c('HEART', 6)]),
    ).toBe('threePairRun');
    expect(
      classifyTienLenCombo([c('SPADE', 4), c('HEART', 4), c('SPADE', 5), c('HEART', 5), c('SPADE', 8), c('HEART', 8)]),
    ).toBe('invalid'); // 4-5 then gap to 8
  });

  it('rejects malformed selections', () => {
    expect(classifyTienLenCombo([c('SPADE', 7), c('HEART', 8)])).toBe('invalid'); // two different ranks
    expect(classifyTienLenCombo([c('SPADE', 3), c('HEART', 4), c('CLOVER', 6)])).toBe('invalid'); // broken run
  });
});

describe('isValidTienLenCombo', () => {
  it('is false for an empty selection and true for a legal combo', () => {
    expect(isValidTienLenCombo([])).toBe(false);
    expect(isValidTienLenCombo([c('SPADE', 7), c('HEART', 8)])).toBe(false);
    expect(isValidTienLenCombo([c('SPADE', 7), c('HEART', 7)])).toBe(true);
  });
});
