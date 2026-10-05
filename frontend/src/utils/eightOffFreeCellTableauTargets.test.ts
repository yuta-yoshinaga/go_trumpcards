import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { eightOffFreeCellTableauTargets } from './eightOffFreeCellTableauTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });

describe('eightOffFreeCellTableauTargets', () => {
  it('returns matching same-suit descending columns and empty columns for Kings only', () => {
    expect(
      eightOffFreeCellTableauTargets(
        [[card('HEART', 12)], [card('SPADE', 12)], [], [card('CLOVER', 13)]],
        card('HEART', 11),
      ),
    ).toEqual([0]);
    expect(eightOffFreeCellTableauTargets([[], [card('DIAMOND', 12)]], card('SPADE', 13))).toEqual([0]);
  });
});
