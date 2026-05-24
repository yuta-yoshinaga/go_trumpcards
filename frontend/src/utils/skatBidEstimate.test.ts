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

  it('extends the "with" run into trump-suit cards when the jack run continues unbroken', () => {
    // Top of clubs is [J♣, J♠, J♥, J♦, A♣, ...]. Holding J♣+J♠+J♥+J♦+A♣ keeps the run
    // unbroken through 5 top trumps → matadors 5, multiplier 6, value 72.
    const hand: Card[] = [c('CLOVER', 11), c('SPADE', 11), c('HEART', 11), c('DIAMOND', 11), c('CLOVER', 1)];
    const estimates = skatBidEstimates(hand);
    const clubs = estimates.find((e) => e.gameType === 'CLOVER');
    if (!clubs) throw new Error('expected CLOVER estimate');
    expect(clubs.matadors).toBe(5);
    expect(clubs.value).toBe(6 * 12);
  });

  it('breaks the "with" run at the first missing top trump', () => {
    // J♣ alone for a clubs game: top[0] = J♣ ✓, top[1] = J♠ ✗ → with 1.
    // A♣ and 10♣ are further down the run but the break at J♠ stops counting.
    const hand: Card[] = [c('CLOVER', 11), c('CLOVER', 1), c('CLOVER', 10)];
    const clubs = skatBidEstimates(hand).find((e) => e.gameType === 'CLOVER');
    if (!clubs) throw new Error('expected CLOVER estimate');
    expect(clubs.matadors).toBe(1);
    expect(clubs.value).toBe(2 * 12);
  });

  it('lets a deep clubs run beat Grand', () => {
    // All 4 jacks + 7 trump clubs (A,10,K,Q,9,8,7) = "with 11" for clubs → 12 × 12 = 144.
    // Grand sees "with 4" jacks → 5 × 24 = 120. Clubs should win.
    const hand: Card[] = [
      c('CLOVER', 11),
      c('SPADE', 11),
      c('HEART', 11),
      c('DIAMOND', 11),
      c('CLOVER', 1),
      c('CLOVER', 10),
      c('CLOVER', 13),
      c('CLOVER', 12),
      c('CLOVER', 9),
      c('CLOVER', 8),
      c('CLOVER', 7),
    ];
    const best = skatBestBidEstimate(hand);
    expect(best.gameType).toBe('CLOVER');
    expect(best.value).toBe(12 * 12);
  });

  it('stops the "without" run at the first held trump (J♠ only ⇒ without 1)', () => {
    // Hand has J♠ but not J♣ → Grand "without 1" (run breaks at J♠), not "without 4".
    const hand: Card[] = [c('SPADE', 11)];
    const estimates = skatBidEstimates(hand);
    const grand = estimates.find((e) => e.gameType === 'GRAND');
    if (!grand) throw new Error('expected GRAND estimate');
    expect(grand.matadors).toBe(1);
    expect(grand.value).toBe(2 * 24);
  });
});
