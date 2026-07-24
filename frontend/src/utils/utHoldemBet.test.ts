import { describe, expect, it } from 'vitest';
import { utHoldemBetBounds } from './utHoldemBet';

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
