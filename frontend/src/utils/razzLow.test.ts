import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { formatRazzLow, razzBestLow } from './razzLow';

const c = (value: number): Card => ({ design: 'SPADE', value });

describe('razzBestLow', () => {
  it('picks the five lowest distinct ranks (Ace low) and formats high-to-low', () => {
    const low = razzBestLow([c(8), c(6), c(4), c(3), c(1), c(13)]);
    expect(low.complete).toBe(true);
    expect(low.ranks).toEqual([1, 3, 4, 6, 8]);
    expect(formatRazzLow(low)).toBe('8-6-4-3-A');
  });

  it('ignores pairs (duplicate ranks do not help)', () => {
    const low = razzBestLow([c(2), c(2), c(5), c(5), c(9)]);
    expect(low.ranks).toEqual([2, 5, 9]);
    expect(low.complete).toBe(false);
  });

  it('marks an incomplete low when fewer than five distinct ranks are available', () => {
    expect(razzBestLow([c(3), c(7), c(10)]).complete).toBe(false);
  });
});
