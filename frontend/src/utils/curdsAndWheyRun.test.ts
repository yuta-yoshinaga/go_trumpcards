import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { isGrabbable, isRunLink, movableFromIndex } from './curdsAndWheyRun';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('isRunLink', () => {
  it('is true for a same-suit descending pair', () => {
    expect(isRunLink(card('SPADE', 9), card('SPADE', 8))).toBe(true);
  });

  it('is false for a different suit', () => {
    expect(isRunLink(card('SPADE', 9), card('HEART', 8))).toBe(false);
  });

  it('is false for a non-descending value', () => {
    expect(isRunLink(card('SPADE', 9), card('SPADE', 7))).toBe(false);
    expect(isRunLink(card('SPADE', 9), card('SPADE', 9))).toBe(false);
  });
});

describe('movableFromIndex', () => {
  it('returns 0 for an empty column', () => {
    expect(movableFromIndex([])).toBe(0);
  });

  it('returns the last index for a single card', () => {
    expect(movableFromIndex([card('SPADE', 5)])).toBe(0);
  });

  it('returns 0 when the whole column is one run', () => {
    expect(movableFromIndex([card('SPADE', 9), card('SPADE', 8), card('SPADE', 7)])).toBe(0);
  });

  it('returns the index where the tail run begins', () => {
    // ♠9, ♥6, ♥5, ♥4 → only ♥6..♥4 form a run → boundary at index 1.
    const col = [card('SPADE', 9), card('HEART', 6), card('HEART', 5), card('HEART', 4)];
    expect(movableFromIndex(col)).toBe(1);
  });

  it('breaks the run at a suit mismatch mid-column', () => {
    // ♥8, ♠7, ♠6 → run only ♠7..♠6 → boundary at index 1.
    const col = [card('HEART', 8), card('SPADE', 7), card('SPADE', 6)];
    expect(movableFromIndex(col)).toBe(1);
  });

  it('returns the last index when only the bottom card is free', () => {
    // ♠5, ♠5 (value gap broken) → boundary at the bottom card.
    const col = [card('SPADE', 5), card('SPADE', 5)];
    expect(movableFromIndex(col)).toBe(1);
  });
});

describe('isGrabbable', () => {
  const col = [card('SPADE', 9), card('HEART', 6), card('HEART', 5), card('HEART', 4)];

  it('is false above the run boundary', () => {
    expect(isGrabbable(col, 0)).toBe(false);
  });

  it('is true at and below the run boundary', () => {
    expect(isGrabbable(col, 1)).toBe(true);
    expect(isGrabbable(col, 2)).toBe(true);
    expect(isGrabbable(col, 3)).toBe(true);
  });

  it('is false for an empty column', () => {
    expect(isGrabbable([], 0)).toBe(false);
  });
});
