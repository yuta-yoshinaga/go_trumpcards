import { describe, expect, it } from 'vitest';
import type { RookPlayerData } from '../types/card';
import { rookBidStatus } from './rookBidStatus';

const p = (id: number, isHuman: boolean, passed: boolean): RookPlayerData =>
  ({ id, isHuman, passed }) as RookPlayerData;

describe('rookBidStatus', () => {
  it('counts every player as active when nobody has passed', () => {
    const status = rookBidStatus([p(0, true, false), p(1, false, false), p(2, false, false), p(3, false, false)]);
    expect(status.activeBidders).toBe(4);
    expect(status.passed).toEqual([]);
  });

  it('lists passed players and reduces the active count', () => {
    const status = rookBidStatus([p(0, true, false), p(1, false, true), p(2, false, false), p(3, false, true)]);
    expect(status.activeBidders).toBe(2);
    expect(status.passed).toEqual([
      { id: 1, isHuman: false },
      { id: 3, isHuman: false },
    ]);
  });

  it('handles all players having passed', () => {
    const status = rookBidStatus([p(0, true, true), p(1, false, true)]);
    expect(status.activeBidders).toBe(0);
    expect(status.passed).toHaveLength(2);
  });
});
