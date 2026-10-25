import { describe, expect, it } from 'vitest';
import {
  anteMax,
  canPlay,
  maxTrips,
  UTH_ANTE_MIN,
  UTH_ANTE_STEP,
  UTH_TRIPS_STEP,
  utHoldemBetBounds,
} from './utHoldemBet';

describe('utHoldemBetBounds', () => {
  it('counts the ante twice (ante + matching blind) plus trips in the total', () => {
    const { total } = utHoldemBetBounds(100, 50, 1000);
    expect(total).toBe(250);
  });

  it('reports the bet as valid when ante*2 + trips fits within chips', () => {
    expect(utHoldemBetBounds(100, 300, 1000).valid).toBe(true);
  });

  it('reports the bet as valid at the exact chip boundary', () => {
    // 500*2 + 0 === 1000
    expect(utHoldemBetBounds(500, 0, 1000).valid).toBe(true);
  });

  it('reports the bet as invalid when ante*2 + trips exceeds chips', () => {
    // Issue example: chips=1000, ante 500 (×2 = 1000) + trips 300 = 1300 > 1000
    expect(utHoldemBetBounds(500, 300, 1000).valid).toBe(false);
  });

  it('caps maxTrips at the chips left after the ante+blind commitment', () => {
    expect(utHoldemBetBounds(300, 0, 1000).maxTrips).toBe(400);
  });

  it('never returns a negative maxTrips when the ante alone exceeds chips', () => {
    expect(utHoldemBetBounds(600, 0, 1000).maxTrips).toBe(0);
  });
});

describe('Ultimate Texas Hold’em shared rules', () => {
  it('exports the Ante and Trips minimum increments and chip based limits', () => {
    expect([UTH_ANTE_MIN, UTH_ANTE_STEP, UTH_TRIPS_STEP]).toEqual([10, 10, 10]);
    expect(anteMax(999)).toBe(499);
    expect(maxTrips(100, 500)).toBe(300);
  });

  it('checks affordability for each Play multiplier', () => {
    expect(canPlay(3, 299, 100)).toBe(false);
    expect(canPlay(3, 300, 100)).toBe(true);
  });
});
