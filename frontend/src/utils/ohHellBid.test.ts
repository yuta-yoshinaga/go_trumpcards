import { describe, expect, it } from 'vitest';
import { ohHellBidSummary } from './ohHellBid';

describe('ohHellBidSummary', () => {
  it('sums placed bids and flags an overbid table', () => {
    expect(ohHellBidSummary([2, 2, 1], 4)).toEqual({ total: 5, diff: 1, kind: 'over' });
  });

  it('flags an underbid table', () => {
    expect(ohHellBidSummary([1, 0], 4)).toEqual({ total: 1, diff: -3, kind: 'under' });
  });

  it('flags an exact table (total equals hand size)', () => {
    expect(ohHellBidSummary([2, 2], 4)).toEqual({ total: 4, diff: 0, kind: 'exact' });
  });

  it('treats an empty bid list as fully under', () => {
    expect(ohHellBidSummary([], 3)).toEqual({ total: 0, diff: -3, kind: 'under' });
  });
});
