import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { eightOffTableauTargets } from './eightOffTableauTargets';

const c = (design: CardDesign, value: number): Card => ({ design, value });

describe('eightOffTableauTargets', () => {
  it('returns only legal same-suit descending and king-to-empty destinations within the supermove limit', () => {
    const tableau = [[c('HEART', 12), c('HEART', 11)], [c('HEART', 12)], [c('SPADE', 13)], [], [], [], [], []];
    expect(eightOffTableauTargets(tableau, [null], 0, 1)).toEqual([1]);
    expect(eightOffTableauTargets(tableau, [null], 0, 0)).toEqual([]);
    expect(eightOffTableauTargets(tableau, [null], 2, 0)).toEqual([3, 4, 5, 6, 7]);
  });

  it('enforces the supermove limit when cells and columns are full, and allows a stack at capacity', () => {
    const source = [c('HEART', 12), c('HEART', 11)];
    const tableau = [
      source,
      [c('HEART', 13)],
      [c('CLOVER', 8)],
      [c('DIAMOND', 7)],
      [c('SPADE', 6)],
      [c('CLOVER', 5)],
      [c('DIAMOND', 4)],
      [c('SPADE', 3)],
    ];

    expect(
      eightOffTableauTargets(
        tableau,
        Array.from({ length: 8 }, () => c('CLOVER', 1)),
        0,
        0,
      ),
    ).toEqual([]);
    expect(eightOffTableauTargets(tableau, [null], 0, 0)).toEqual([1]);
  });

  it('rejects invalid source sequences and out-of-range selections', () => {
    expect(eightOffTableauTargets([[c('HEART', 12), c('SPADE', 11)], [], [], [], [], [], [], []], [], 0, 0)).toEqual(
      [],
    );
    expect(eightOffTableauTargets([[], [], [], [], [], [], [], []], [], 0, 0)).toEqual([]);
  });
});
