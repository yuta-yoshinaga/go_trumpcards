import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { skatBestBidEstimate, skatBidEstimates } from './skatBidEstimate';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('skatBidEstimates', () => {
  it('returns base value for an empty hand (without ∞ jacks, but capped by run)', () => {
    const estimates = skatBidEstimates([]);
    // Empty hand → "without" run = full set of top trumps. For Grand without 4 jacks = matadors 4, multiplier 5, value 120.
    const grand = estimates.find((e) => e.gameType === 'GRAND');
    if (!grand) throw new Error('expected GRAND estimate');
    expect(grand.matadors).toBe(4);
    expect(grand.value).toBe(5 * 24);
  });

  it('handles a single J♣ as "with 1" for clubs and grand', () => {
    const hand: Card[] = [c('CLOVER', 11)];
    const estimates = skatBidEstimates(hand);
    const grand = estimates.find((e) => e.gameType === 'GRAND');
    if (!grand) throw new Error('expected GRAND estimate');
    expect(grand.matadors).toBe(1);
    expect(grand.value).toBe(2 * 24);
  });

  it('skatBestBidEstimate picks the highest game value', () => {
    const hand: Card[] = [c('CLOVER', 11), c('SPADE', 11)];
    const best = skatBestBidEstimate(hand);
    // With 2 jacks, Grand = 3 × 24 = 72; Clubs suit = 3 × 12 = 36. Grand wins.
    expect(best.gameType).toBe('GRAND');
    expect(best.value).toBe(72);
  });
});
