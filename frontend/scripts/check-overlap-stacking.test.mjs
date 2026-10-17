import { describe, expect, it } from 'vitest';
import { findOverlapStackingViolations } from './check-overlap-stacking.mjs';

describe('check-overlap-stacking', () => {
  it('allows positioned opacity cards with negative inline margin', () => {
    expect(
      findOverlapStackingViolations(
        '<button className="relative disabled:opacity-50" style={{ marginLeft: idx > 0 ? -Math.round(width) : undefined }} />',
      ),
    ).toEqual([]);
  });
  it('reports an unpositioned opacity card with negative margin', () => {
    expect(
      findOverlapStackingViolations(
        '<button className="disabled:opacity-50" style={{ marginLeft: idx > 0 ? -Math.round(width) : undefined }} />',
      ),
    ).toHaveLength(1);
  });
  it('allows overlapping elements without opacity classes', () => {
    expect(findOverlapStackingViolations('<button className="" style={{ marginLeft: -8 }} />')).toEqual([]);
  });
  it('allows opacity elements without overlap', () => {
    expect(findOverlapStackingViolations('<button className="disabled:opacity-50" />')).toEqual([]);
  });
  it('checks opacity children of a negative space-x parent', () => {
    expect(
      findOverlapStackingViolations(
        '<div className="flex -space-x-2"><button className="disabled:opacity-50" /></div>',
      ),
    ).toHaveLength(1);
  });
});
