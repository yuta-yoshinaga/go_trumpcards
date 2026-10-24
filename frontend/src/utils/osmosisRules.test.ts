import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { osmosisAllowedRanks } from './osmosisRules';

const c = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

describe('osmosisAllowedRanks', () => {
  it('an empty base row (0) accepts only the base rank', () => {
    expect(osmosisAllowedRanks([[], [], [], []], 5, 0)).toEqual([5]);
  });

  it('an empty row >=1 accepts the base rank only when the row above is non-empty', () => {
    expect(osmosisAllowedRanks([[], [], [], []], 5, 1)).toEqual([]);
    expect(osmosisAllowedRanks([[c('SPADE', 5)], [], [], []], 5, 1)).toEqual([5]);
  });

  it('the base row with a fixed suit accepts any not-yet-placed rank', () => {
    const allowed = osmosisAllowedRanks([[c('SPADE', 5), c('SPADE', 6)], [], [], []], 5, 0);
    expect(allowed).not.toContain(5);
    expect(allowed).not.toContain(6);
    expect(allowed).toContain(1);
    expect(allowed).toContain(13);
    expect(allowed).toHaveLength(11);
  });

  it('row >=1 accepts only ranks present in the row above, minus those already placed', () => {
    const foundation = [
      [c('SPADE', 5), c('SPADE', 6), c('SPADE', 7)], // base row has 5,6,7
      [c('HEART', 5)], // row 1 already has 5
      [],
      [],
    ];
    // Row 1 can still add 6 and 7 (present above, not yet in row 1).
    expect(osmosisAllowedRanks(foundation, 5, 1)).toEqual([6, 7]);
  });
});
