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

  it('accepts only adjacent ranks on the waste for a quilt card', () => {
    expect(crazyQuiltLegalTargets([], undefined, card('HEART', 6), card('SPADE', 7), true).waste).toBe(true);
    expect(crazyQuiltLegalTargets([], undefined, card('HEART', 6), card('SPADE', 8), true).waste).toBe(false);
  });
});
