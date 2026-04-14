import { describe, expect, it } from 'vitest';
import { isAdjacentRank } from './speedUtils';

describe('isAdjacentRank', () => {
  it('returns true for adjacent ranks', () => {
    expect(isAdjacentRank(1, 2)).toBe(true);
    expect(isAdjacentRank(5, 4)).toBe(true);
    expect(isAdjacentRank(12, 13)).toBe(true);
  });

  it('returns true for K↔A wrap-around', () => {
    expect(isAdjacentRank(1, 13)).toBe(true);
    expect(isAdjacentRank(13, 1)).toBe(true);
  });

  it('returns false for non-adjacent ranks', () => {
    expect(isAdjacentRank(1, 3)).toBe(false);
    expect(isAdjacentRank(5, 10)).toBe(false);
    expect(isAdjacentRank(7, 7)).toBe(false);
  });
});
