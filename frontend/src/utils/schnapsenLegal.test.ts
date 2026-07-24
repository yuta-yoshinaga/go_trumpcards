import { describe, expect, it } from 'vitest';
import { computeSchnapsenLegalRing } from './schnapsenLegal';

describe('computeSchnapsenLegalRing', () => {
  it('returns an empty set in phase 1 (open talon, any card legal)', () => {
    expect(computeSchnapsenLegalRing(false, true, [0, 1, 2]).size).toBe(0);
  });

  it('returns an empty set when it is not the human turn', () => {
    expect(computeSchnapsenLegalRing(true, false, [0, 1]).size).toBe(0);
  });

  it('returns the backend valid plays as a set in phase 2 on the human turn', () => {
    const ring = computeSchnapsenLegalRing(true, true, [1, 3]);
    expect([...ring].sort()).toEqual([1, 3]);
    expect(ring.has(1)).toBe(true);
    expect(ring.has(0)).toBe(false);
  });

  it('treats undefined validPlays as an empty set', () => {
    expect(computeSchnapsenLegalRing(true, true, undefined).size).toBe(0);
  });
});
