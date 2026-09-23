import { describe, expect, it } from 'vitest';
import { bisleyNextRank } from './bisleyNextRank';

describe('bisleyNextRank', () => {
  it('returns null when ascending and descending foundations meet', () => {
    expect(bisleyNextRank(6, 13, 'ascending')).toBeNull();
    expect(bisleyNextRank(7, 13, 'descending')).toBeNull();
  });

  it('returns the next rank in both directions when one card remains', () => {
    expect(bisleyNextRank(6, 12, 'ascending')).toBe(7);
    expect(bisleyNextRank(7, 12, 'descending')).toBe(6);
  });

  it('returns the next ascending rank and stops after the King', () => {
    expect(bisleyNextRank(undefined, 0, 'ascending')).toBe(1);
    expect(bisleyNextRank(5, 5, 'ascending')).toBe(6);
    expect(bisleyNextRank(13, 13, 'ascending')).toBeNull();
  });

  it('returns the next descending rank and stops after the Ace', () => {
    expect(bisleyNextRank(undefined, 0, 'descending')).toBe(13);
    expect(bisleyNextRank(10, 3, 'descending')).toBe(9);
    expect(bisleyNextRank(1, 13, 'descending')).toBeNull();
  });
});
