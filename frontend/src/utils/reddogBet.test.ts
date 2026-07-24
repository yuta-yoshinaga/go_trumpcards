import { describe, expect, it } from 'vitest';
import { canRedDogRaise, REDDOG_MIN_BET } from './reddogBet';

describe('reddogBet', () => {
  it('exposes the backend minimum bet value', () => {
    expect(REDDOG_MIN_BET).toBe(10);
  });

  it('allows a raise when chips meet or exceed the minimum bet', () => {
    expect(canRedDogRaise(REDDOG_MIN_BET)).toBe(true);
    expect(canRedDogRaise(REDDOG_MIN_BET + 1)).toBe(true);
    expect(canRedDogRaise(1000)).toBe(true);
  });

  it('blocks a raise when chips fall below the minimum bet', () => {
    expect(canRedDogRaise(REDDOG_MIN_BET - 1)).toBe(false);
    expect(canRedDogRaise(0)).toBe(false);
  });
});
