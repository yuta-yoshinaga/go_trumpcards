import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { scoponeSelectionSum } from './scoponeSelectionSum';

const c = (design: CardDesign, value: number): Card => ({ design, value });
const table = [c('SPADE', 3), c('HEART', 4), c('CLOVER', 7)];

describe('scoponeSelectionSum', () => {
  it('is null with no hand card chosen', () => {
    expect(scoponeSelectionSum(null, table, [0])).toBeNull();
    expect(scoponeSelectionSum(undefined, table, [0])).toBeNull();
  });

  it('starts at zero against the chosen card once one is picked', () => {
    expect(scoponeSelectionSum(c('SPADE', 7), table, [])).toEqual({ sum: 0, target: 7 });
  });

  it('adds up the selected table cards', () => {
    expect(scoponeSelectionSum(c('SPADE', 7), table, [0, 1])).toEqual({ sum: 7, target: 7 });
  });

  it('can overshoot the target', () => {
    expect(scoponeSelectionSum(c('SPADE', 4), table, [1, 2])).toEqual({ sum: 11, target: 4 });
  });

  it('ignores an index that is not on the table', () => {
    expect(scoponeSelectionSum(c('SPADE', 7), table, [0, 99])).toEqual({ sum: 3, target: 7 });
  });
});
