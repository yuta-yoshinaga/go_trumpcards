import { describe, expect, it } from 'vitest';
import { formatSignedDelta } from './formatSignedDelta';

describe('formatSignedDelta', () => {
  it('prefixes positive numbers with a plus sign', () => {
    expect(formatSignedDelta(10)).toBe('+10');
    expect(formatSignedDelta(1)).toBe('+1');
  });

  it('formats zero as ±0', () => {
    expect(formatSignedDelta(0)).toBe('±0');
  });

  it('formats negative numbers with standard minus sign', () => {
    expect(formatSignedDelta(-5)).toBe('-5');
    expect(formatSignedDelta(-1)).toBe('-1');
  });
});
