import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { scopaCaptureValue, scopaTakeCandidates } from './scopaTakeCandidates';

const card = (value: number, design = 'SPADE'): Card => ({ design, value }) as unknown as Card;

describe('scopaCaptureValue', () => {
  it('maps number cards to their face value', () => {
    expect(scopaCaptureValue(1)).toBe(1);
    expect(scopaCaptureValue(7)).toBe(7);
    expect(scopaCaptureValue(10)).toBe(10);
  });

  it('maps face cards: J=8, Q=9, K=10', () => {
    expect(scopaCaptureValue(11)).toBe(8);
    expect(scopaCaptureValue(12)).toBe(9);
    expect(scopaCaptureValue(13)).toBe(10);
  });

  it('returns 0 for out-of-range values', () => {
    expect(scopaCaptureValue(0)).toBe(0);
    expect(scopaCaptureValue(14)).toBe(0);
  });
});

describe('scopaTakeCandidates', () => {
  it('returns empty for empty table', () => {
    expect(scopaTakeCandidates([], 5).indices.size).toBe(0);
  });

  it('returns empty for non-positive target', () => {
    expect(scopaTakeCandidates([card(5)], 0).indices.size).toBe(0);
  });

  it('forces single match when a single card equals the target', () => {
    // table: [5, 2, 3]; target 5 → only the single 5 (index 0), not 2+3
    const { indices } = scopaTakeCandidates([card(5), card(2), card(3)], 5);
    expect([...indices].sort()).toEqual([0]);
  });

  it('returns all matching singles when several equal the target', () => {
    const { indices } = scopaTakeCandidates([card(5), card(5), card(3)], 5);
    expect([...indices].sort()).toEqual([0, 1]);
  });

  it('enumerates subset sums when no single matches', () => {
    // table: [2, 3, 4]; target 5 → subset {2,3} (indices 0,1)
    const { indices } = scopaTakeCandidates([card(2), card(3), card(4)], 5);
    expect([...indices].sort()).toEqual([0, 1]);
  });

  it('includes every index touched by any matching subset', () => {
    // table: [1, 4, 2, 3]; target 5 → {1,4} and {2,3} → all 4 indices
    const { indices } = scopaTakeCandidates([card(1), card(4), card(2), card(3)], 5);
    expect([...indices].sort()).toEqual([0, 1, 2, 3]);
  });

  it('uses capture values for face cards on the table', () => {
    // King = 10; target 10 → single match on the King at index 0
    const { indices } = scopaTakeCandidates([card(13), card(6), card(4)], 10);
    expect([...indices].sort()).toEqual([0]);
  });

  it('returns empty when no capture is possible', () => {
    const { indices } = scopaTakeCandidates([card(8), card(9)], 5);
    expect(indices.size).toBe(0);
  });
});
