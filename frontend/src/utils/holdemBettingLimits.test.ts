import { describe, expect, it } from 'vitest';
import { getHoldemBettingLimitName } from './holdemBettingLimits';

describe('getHoldemBettingLimitName', () => {
  it.each([
    [0, 'Fixed'],
    [1, 'Pot Limit'],
    [2, 'No Limit'],
  ])('returns %s for betting-limit index %s', (index, expectedName) => {
    expect(getHoldemBettingLimitName(index)).toBe(expectedName);
  });

  it.each([99, -1])('returns Unknown for out-of-range index %s', (index) => {
    expect(getHoldemBettingLimitName(index)).toBe('Unknown');
  });
});
