import { describe, expect, it } from 'vitest';
import { computeOverlap, rowCountFor, stackedRowMarginTop, TWO_ROW_THRESHOLD } from './handOverlap';

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

describe('rowCountFor', () => {
  it('keeps fewer than the threshold in one row', () => {
    expect(rowCountFor(TWO_ROW_THRESHOLD - 1, 46, 375)).toBe(1);
  });

  it('starts at two rows when the threshold is reached', () => {
    expect(rowCountFor(TWO_ROW_THRESHOLD, 46, 375)).toBe(2);
  });

  it('allows one row when the cards fit and single rows are allowed', () => {
    expect(rowCountFor(5, 46, 375, { allowSingleRow: true })).toBe(1);
  });

  it('keeps multiple rows when cards do not fit, even when single rows are allowed', () => {
    expect(rowCountFor(5, 200, 375, { allowSingleRow: true })).toBe(rowCountFor(5, 200, 375));
  });

  it('keeps the two-row default when single rows are not allowed', () => {
    expect(rowCountFor(5, 46, 375)).toBe(2);
  });

  it('adds rows when the largest row still requires scrolling', () => {
    expect(rowCountFor(20, 106, 375)).toBeGreaterThanOrEqual(3);
  });
});

describe('stackedRowMarginTop', () => {
  it('leaves the configured visible part of the preceding row', () => {
    expect(stackedRowMarginTop(40)).toBe(-14);
  });
});
