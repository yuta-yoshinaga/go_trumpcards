import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { pitchHandPipBreakdown, pitchHandPips, pitchPipValue } from './pitchUtils';

const c = (value: number, design: Card['design'] = 'SPADE'): Card => ({ design, value });

describe('pitchPipValue', () => {
  it.each([
    [1, 4],
    [10, 10],
    [11, 1],
    [12, 2],
    [13, 3],
  ])('value %i scores %i pip(s)', (value, expected) => {
    expect(pitchPipValue(value)).toBe(expected);
  });

  it.each([2, 3, 4, 5, 6, 7, 8, 9])('value %i scores 0 pips', (value) => {
    expect(pitchPipValue(value)).toBe(0);
  });
});

describe('pitchHandPips', () => {
  it('returns 0 for an empty hand', () => {
    expect(pitchHandPips([])).toBe(0);
  });

  it('sums pip values across the hand', () => {
    // A(4) + 10(10) + J(1) + 7(0) = 15
    expect(pitchHandPips([c(1), c(10), c(11), c(7)])).toBe(15);
  });

  it('includes K and Q in the total', () => {
    // K(3) + Q(2) + 5(0) = 5
    expect(pitchHandPips([c(13), c(12), c(5)])).toBe(5);
  });
});

describe('pitchHandPipBreakdown', () => {
  it('returns the per-card breakdown in hand order', () => {
    expect(pitchHandPipBreakdown([c(1), c(10), c(7)])).toEqual([
      { value: 1, pips: 4 },
      { value: 10, pips: 10 },
      { value: 7, pips: 0 },
    ]);
  });
});
