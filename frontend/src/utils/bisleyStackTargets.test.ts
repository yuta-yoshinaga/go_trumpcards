import { describe, expect, it } from 'vitest';
import type { Card } from '../types/common';
import { getBisleyStackTargets } from './bisleyStackTargets';

describe('getBisleyStackTargets', () => {
  it('returns empty array for an empty column (null)', () => {
    expect(getBisleyStackTargets(null)).toEqual([]);
  });

  it('returns [2] for an Ace (value 1)', () => {
    const card: Card = { design: 'SPADE', value: 1 };
    expect(getBisleyStackTargets(card)).toEqual([2]);
  });

  it('returns [12] for a King (value 13)', () => {
    const card: Card = { design: 'SPADE', value: 13 };
    expect(getBisleyStackTargets(card)).toEqual([12]);
  });

  it('returns [4, 6] for a 5', () => {
    const card: Card = { design: 'SPADE', value: 5 };
    expect(getBisleyStackTargets(card)).toEqual([4, 6]);
  });

  it('returns [9, 11] for a 10', () => {
    const card: Card = { design: 'HEART', value: 10 };
    expect(getBisleyStackTargets(card)).toEqual([9, 11]);
  });
});
