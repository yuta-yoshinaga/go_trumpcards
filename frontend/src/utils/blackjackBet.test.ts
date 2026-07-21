import { describe, expect, it } from 'vitest';
import { BJ_CHIP_VALUES, BJ_MIN_BET, bjAddChip, bjQuickBetAmount } from './blackjackBet';

describe('bjQuickBetAmount', () => {
  it('returns the table minimum for "min"', () => {
    expect(bjQuickBetAmount('min', 1000)).toBe(BJ_MIN_BET);
    expect(bjQuickBetAmount('min', 30)).toBe(BJ_MIN_BET);
  });

  it('returns half the chips rounded down to a 10-multiple for "half"', () => {
    expect(bjQuickBetAmount('half', 1000)).toBe(500);
    expect(bjQuickBetAmount('half', 250)).toBe(120); // floor(125/10)*10
  });

  it('returns all chips rounded down to a 10-multiple for "max"', () => {
    expect(bjQuickBetAmount('max', 1000)).toBe(1000);
    expect(bjQuickBetAmount('max', 255)).toBe(250);
  });

  it('never exceeds the available chips, including below the table minimum', () => {
    for (const kind of ['min', 'half', 'max'] as const) {
      expect(bjQuickBetAmount(kind, 5)).toBeLessThanOrEqual(5);
      expect(bjQuickBetAmount(kind, 15)).toBeLessThanOrEqual(15);
    }
  });
});

describe('bjAddChip', () => {
  it('adds the chip value to the current bet', () => {
    expect(bjAddChip(10, 25, 1000)).toBe(35);
    expect(bjAddChip(100, 100, 1000)).toBe(200);
  });

  it('clamps the result to the available chips', () => {
    expect(bjAddChip(90, 25, 100)).toBe(100);
    expect(bjAddChip(500, 100, 500)).toBe(500);
  });

  it('offers the +10 / +25 / +100 denominations', () => {
    expect(BJ_CHIP_VALUES).toEqual([10, 25, 100]);
  });
});
