import { describe, expect, it } from 'vitest';
import { computeOverlap } from './handOverlap';

describe('computeOverlap', () => {
  it('uses a small positive gap when cards fit without overlap', () => {
    expect(computeOverlap(3, 46, 375)).toEqual({ overlap: 2, useScroll: false });
  });

  it('returns a negative overlap when cards fit with overlap', () => {
    const result = computeOverlap(7, 46, 343);
    expect(result.overlap).toBeLessThan(0);
    expect(result.useScroll).toBe(false);
  });

  it('requests scrolling when minimum card exposure cannot fit', () => {
    expect(computeOverlap(13, 106, 375).useScroll).toBe(true);
  });
});
