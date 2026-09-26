import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { crazyQuiltLegalTargets } from './crazyQuiltLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });

describe('crazyQuiltLegalTargets', () => {
  it('accepts the next same-suit rank on an ascending foundation', () => {
    const targets = crazyQuiltLegalTargets([[card('SPADE', 4)]], [true], null, card('SPADE', 5), false);
    expect(targets.foundation).toEqual([true]);
  });

  it('rejects an empty foundation', () => {
    const targets = crazyQuiltLegalTargets([[]], [true], null, card('SPADE', 1), false);
    expect(targets.foundation).toEqual([false]);
  });

  it('accepts the next same-suit rank on a descending foundation', () => {
    const targets = crazyQuiltLegalTargets([[card('SPADE', 5)]], [false], null, card('SPADE', 4), false);
    expect(targets.foundation).toEqual([true]);
  });

  it('rejects a wrong suit, a non-adjacent rank, a full foundation, and no selected card', () => {
    const targets = crazyQuiltLegalTargets(
      [[card('SPADE', 4)], [card('HEART', 4)], Array.from({ length: 13 }, () => card('DIAMOND', 7))],
      undefined,
      null,
      card('HEART', 6),
      false,
    );
    expect(targets.foundation).toEqual([false, false, false]);
    expect(crazyQuiltLegalTargets([[card('SPADE', 4)]], undefined, null, null, false).foundation).toEqual([false]);
  });

  it('accepts only adjacent ranks on the waste for a quilt card', () => {
    expect(crazyQuiltLegalTargets([], undefined, card('HEART', 6), card('SPADE', 7), true).waste).toBe(true);
    expect(crazyQuiltLegalTargets([], undefined, card('HEART', 8), card('SPADE', 7), true).waste).toBe(true);
    expect(crazyQuiltLegalTargets([], undefined, card('HEART', 6), card('SPADE', 8), true).waste).toBe(false);
    expect(crazyQuiltLegalTargets([], undefined, null, card('SPADE', 7), true).waste).toBe(false);
    expect(crazyQuiltLegalTargets([], undefined, card('HEART', 6), card('SPADE', 7), false).waste).toBe(false);
    expect(crazyQuiltLegalTargets([], undefined, card('HEART', 6), null, true).waste).toBe(false);
  });
});
