import { describe, expect, it } from 'vitest';
import { computeLooPotRisk } from './looPotRisk';

describe('computeLooPotRisk', () => {
  it('derives the loo penalty from potStart and the per-trick share from potStart/5', () => {
    // Default deal: 4 players ante 3 → pot 12, potStart 12.
    expect(computeLooPotRisk(12, 12)).toEqual({ pot: 12, looPenalty: 12, perTrick: 2, maxWin: 10 });
  });

  it('floors the per-trick share when potStart is not divisible by 5', () => {
    // potStart 18 → 18/5 = 3.6 → floored to 3; penalty stays the full 18.
    expect(computeLooPotRisk(18, 18)).toEqual({ pot: 18, looPenalty: 18, perTrick: 3, maxWin: 15 });
  });

  it('keeps the shown pot independent of the potStart-based risk figures', () => {
    // Pot can differ from potStart mid-deal; the penalty always tracks potStart.
    expect(computeLooPotRisk(20, 15)).toEqual({ pot: 20, looPenalty: 15, perTrick: 3, maxWin: 15 });
  });

  // **全トリック取っても pot 全額は入らない。**端数はポットに残る。
  it('reports the reachable maximum, which is below the pot when it does not divide by five', () => {
    const { pot, maxWin } = computeLooPotRisk(37, 37);
    expect(maxWin).toBe(35);
    expect(maxWin).toBeLessThan(pot);
  });

  it('clamps negative inputs to zero', () => {
    expect(computeLooPotRisk(-5, -5)).toEqual({ pot: 0, looPenalty: 0, perTrick: 0, maxWin: 0 });
  });

  it('yields a zero per-trick share for a pot smaller than one share', () => {
    expect(computeLooPotRisk(4, 4)).toEqual({ pot: 4, looPenalty: 4, perTrick: 0, maxWin: 0 });
  });
});
