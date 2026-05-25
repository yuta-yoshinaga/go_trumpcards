import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { cassinoTakeCandidates } from './cassinoTakeCandidates';

const c = (value: number): Card => ({ design: 'SPADE', value });

describe('cassinoTakeCandidates', () => {
  it('returns empty for empty table', () => {
    expect(cassinoTakeCandidates([], 9).indices.size).toBe(0);
  });

  it('matches a single equal-value card', () => {
    const r = cassinoTakeCandidates([c(2), c(9), c(5)], 9);
    expect(Array.from(r.indices).sort()).toEqual([1]);
  });

  it('matches a pair that sums to the target', () => {
    const r = cassinoTakeCandidates([c(4), c(5), c(2)], 9);
    expect(Array.from(r.indices).sort()).toEqual([0, 1]);
  });

  it('returns the union of every matching subset', () => {
    // table: 2, 3, 5, 9 target 9. Matches: {9}, {4? no}, {2,3,? no}, {2,3, ? }. Actually 2+3+? hmm 2+3=5,+nothing.
    // Match subsets: {9} → index 3; {4? none}. Try {2,3, 4? none}. Only {9}. But add a pair: 4+5=9.
    const r = cassinoTakeCandidates([c(2), c(3), c(4), c(5), c(9)], 9);
    // Matches: {9}(idx4); {4,5}(idx 2,3); {2,3,4}(idx 0,1,2).
    expect(Array.from(r.indices).sort()).toEqual([0, 1, 2, 3, 4]);
  });

  it('restricts to identical-value matches for face cards (J/Q/K)', () => {
    const r = cassinoTakeCandidates([c(11), c(5), c(6), c(11)], 11);
    expect(Array.from(r.indices).sort()).toEqual([0, 3]);
  });

  it('returns empty when no subset sums to target', () => {
    const r = cassinoTakeCandidates([c(2), c(2), c(3)], 9);
    expect(r.indices.size).toBe(0);
  });
});
