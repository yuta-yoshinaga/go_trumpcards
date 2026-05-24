import { describe, expect, it } from 'vitest';
import { computeIndianPokerEquity } from './indianPokerEquity';

describe('computeIndianPokerEquity', () => {
  it('returns 1 when there are no opponents', () => {
    expect(computeIndianPokerEquity([])).toBe(1);
  });

  it('returns 0 when the only opponent shows a King (no rank above)', () => {
    // Max opp = 13; ranks above = 0; remaining = 51; equity = 0/51 = 0.
    expect(computeIndianPokerEquity([13])).toBe(0);
  });

  it('returns a high equity when the only opponent shows a 2', () => {
    // Max opp = 2; ranks above = 11 ranks × 4 = 44; remaining = 51.
    const eq = computeIndianPokerEquity([2]);
    expect(eq).toBeCloseTo(44 / 51, 6);
  });

  it('subtracts visible cards in the above-max range', () => {
    // Three opponents: 10, 11, 12. Max = 12. Above = K(4 copies) only.
    // No visible above-max cards, so above = 4. Remaining = 52 - 3 = 49.
    expect(computeIndianPokerEquity([10, 11, 12])).toBeCloseTo(4 / 49, 6);
  });

  it('ignores invalid rank values', () => {
    expect(computeIndianPokerEquity([0, 14, NaN, 5])).toBeCloseTo(
      // Only "5" is valid; max = 5; above = 8 ranks × 4 = 32; remaining = 51.
      32 / 51,
      6,
    );
  });

  it('returns 0 when every remaining card is at or below the opponent max', () => {
    // Two opponents with 13 each (theoretical edge case in larger deck).
    expect(computeIndianPokerEquity([13, 13])).toBe(0);
  });
});
