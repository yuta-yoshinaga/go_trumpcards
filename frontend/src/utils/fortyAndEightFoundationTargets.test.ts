import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { fortyAndEightFoundationTargets } from './fortyAndEightFoundationTargets';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('fortyAndEightFoundationTargets', () => {
  it('returns an empty set when no card is selected', () => {
    expect(fortyAndEightFoundationTargets(null, [[], [], [], [], [], [], [], []]).size).toBe(0);
    expect(fortyAndEightFoundationTargets(undefined, [[]]).size).toBe(0);
  });

  it('marks every empty pile eligible for an Ace', () => {
    const foundation: Card[][] = [[], [], [card('HEART', 2)]];
    const targets = fortyAndEightFoundationTargets(card('SPADE', 1), foundation);
    expect(targets.has(0)).toBe(true);
    expect(targets.has(1)).toBe(true);
    // Non-empty pile does not accept an Ace.
    expect(targets.has(2)).toBe(false);
  });

  it('does not mark empty piles for a non-Ace', () => {
    const targets = fortyAndEightFoundationTargets(card('SPADE', 5), [[], []]);
    expect(targets.size).toBe(0);
  });

  it('accepts a same-suit card exactly one rank above the pile top', () => {
    const foundation: Card[][] = [[card('SPADE', 1), card('SPADE', 2)], [card('HEART', 1)]];
    const targets = fortyAndEightFoundationTargets(card('SPADE', 3), foundation);
    expect(targets.has(0)).toBe(true);
    expect(targets.has(1)).toBe(false);
  });

  it('rejects a wrong-suit or wrong-rank card', () => {
    const foundation: Card[][] = [[card('SPADE', 1)]];
    expect(fortyAndEightFoundationTargets(card('HEART', 2), foundation).size).toBe(0); // wrong suit
    expect(fortyAndEightFoundationTargets(card('SPADE', 3), foundation).size).toBe(0); // skips a rank
  });

  it('can mark multiple non-empty piles of the same suit', () => {
    const foundation: Card[][] = [[card('SPADE', 1)], [card('SPADE', 1)]];
    const targets = fortyAndEightFoundationTargets(card('SPADE', 2), foundation);
    expect(targets.has(0)).toBe(true);
    expect(targets.has(1)).toBe(true);
  });
});
