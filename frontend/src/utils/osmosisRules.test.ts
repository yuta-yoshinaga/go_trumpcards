import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { osmosisAllowedRanks, osmosisCanPlace } from './osmosisRules';

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

describe('osmosisCanPlace', () => {
  const foundation = [[c('SPADE', 5), c('SPADE', 6)], [c('HEART', 5)], [], []];

  it('base row accepts any rank of its fixed suit', () => {
    expect(osmosisCanPlace(foundation, 5, 0, c('SPADE', 9))).toBe(true);
    expect(osmosisCanPlace(foundation, 5, 0, c('HEART', 9))).toBe(false); // wrong suit
  });

  it('row >=1 accepts a matching-suit card whose rank is in the row above', () => {
    expect(osmosisCanPlace(foundation, 5, 1, c('HEART', 6))).toBe(true); // 6 is in base row
    expect(osmosisCanPlace(foundation, 5, 1, c('HEART', 9))).toBe(false); // 9 not above
    expect(osmosisCanPlace(foundation, 5, 1, c('SPADE', 6))).toBe(false); // wrong suit for row 1
  });

  it('an empty row accepts only the base rank and a fresh suit', () => {
    expect(osmosisCanPlace(foundation, 5, 2, c('CLOVER', 5))).toBe(true); // base rank, new suit, row 1 non-empty
    expect(osmosisCanPlace(foundation, 5, 2, c('CLOVER', 6))).toBe(false); // not base rank
    expect(osmosisCanPlace(foundation, 5, 2, c('SPADE', 5))).toBe(false); // suit already assigned (row 0)
  });
});
