import { describe, expect, it } from 'vitest';
import { computeBouillottePotOdds } from './bouillottePotOdds';

describe('computeBouillottePotOdds', () => {
  it('computes percentage and simplified ratio for a known pot and call', () => {
    // Human already wagered 10, must reach 20 -> call 10 into a pot of 40.
    // Pot odds: 10 / (40 + 10) = 20%, ratio 40:10 -> 4:1.
    const odds = computeBouillottePotOdds(40, 20, 10);
    expect(odds).toEqual({
      callAmount: 10,
      isFree: false,
      percentage: 20,
      ratioPot: 4,
      ratioCall: 1,
    });
  });

  it('rounds the percentage to one decimal place', () => {
    // call 10 into pot 20 -> 10 / 30 = 33.333...% -> 33.3%.
    const odds = computeBouillottePotOdds(20, 20, 10);
    expect(odds.percentage).toBe(33.3);
    expect(odds.ratioPot).toBe(2);
    expect(odds.ratioCall).toBe(1);
  });

  it('reports a free check when nothing is owed (call amount 0)', () => {
    const odds = computeBouillottePotOdds(40, 10, 10);
    expect(odds).toEqual({
      callAmount: 0,
      isFree: true,
      percentage: 0,
      ratioPot: 0,
      ratioCall: 0,
    });
  });

  it('treats an already-matched bet (roundBet > currentBet) as free', () => {
    const odds = computeBouillottePotOdds(40, 10, 20);
    expect(odds.isFree).toBe(true);
    expect(odds.callAmount).toBe(0);
  });

  it('handles an empty pot: call is the whole resulting pot (100%)', () => {
    const odds = computeBouillottePotOdds(0, 10, 0);
    expect(odds.percentage).toBe(100);
    expect(odds.ratioPot).toBe(0);
    expect(odds.ratioCall).toBe(1);
  });
});
