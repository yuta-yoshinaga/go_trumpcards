import { describe, expect, it } from 'vitest';
import { clampThreeCardBragRaise, threeCardBragRaiseBounds } from './threeCardBragRaise';

describe('threeCardBragRaiseBounds', () => {
  it('sets min to stake + 1', () => {
    expect(threeCardBragRaiseBounds(5, 100, false).min).toBe(6);
  });

  it('caps a Blind player at their full chip count', () => {
    const b = threeCardBragRaiseBounds(2, 40, false);
    expect(b.max).toBe(40);
    expect(b.canRaise).toBe(true);
  });

  it('halves the ceiling for a Seen player (they pay double)', () => {
    expect(threeCardBragRaiseBounds(2, 41, true).max).toBe(20);
  });

  it('reports canRaise=false when chips are too low', () => {
    // Seen with 4 chips: max = 2, min = stake + 1 = 3 -> no legal raise.
    const b = threeCardBragRaiseBounds(2, 4, true);
    expect(b.max).toBe(2);
    expect(b.canRaise).toBe(false);
  });
});

describe('clampThreeCardBragRaise', () => {
  it('clamps values below min up to min', () => {
    expect(clampThreeCardBragRaise(1, 3, 50)).toBe(3);
  });

  it('clamps values above max down to max', () => {
    expect(clampThreeCardBragRaise(99, 3, 50)).toBe(50);
  });

  it('leaves in-range values unchanged', () => {
    expect(clampThreeCardBragRaise(10, 3, 50)).toBe(10);
  });

  it('returns min when no legal raise exists (max < min)', () => {
    expect(clampThreeCardBragRaise(10, 5, 2)).toBe(5);
  });
});
