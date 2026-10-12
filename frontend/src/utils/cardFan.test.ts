import { describe, expect, it } from 'vitest';
import { fanOverlap, fanRows } from './cardFan';

describe('card fan layout', () => {
  it('keeps a fitting row at its requested gap', () => {
    expect(fanOverlap(3, 50, 166, 28, 8)).toBe(8);
  });
  it('overlaps cards to fit one row', () => {
    expect(fanOverlap(13, 106, 1008, 28, 8)).toBeLessThan(0);
  });
  it('does not overlap rows with zero or one card', () => {
    expect(fanOverlap(1, 106, 1008, 28, 8)).toBe(0);
    expect(fanOverlap(0, 106, 1008, 28, 8)).toBe(0);
  });
  it('uses multiple rows when exposure requires it', () => {
    expect(fanRows(13, 106, 300, 28)).toBeGreaterThan(1);
  });
  it('fits 13 cards at desktop exposure but needs two rows at touch exposure', () => {
    expect(fanRows(13, 46, 400, 44)).toBe(2);
    expect(fanRows(13, 46, 400, 28)).toBe(1);
  });
  it('handles empty/single hands and zero available width', () => {
    expect(fanRows(0, 50, 100, 28)).toBe(1);
    expect(fanRows(1, 50, 100, 28)).toBe(1);
    expect(fanRows(13, 50, 0, 28)).toBe(1);
  });
  it('clamps overlap to minimum exposure', () => {
    expect(fanOverlap(13, 106, 50, 28, 8)).toBe(-78);
  });
});
