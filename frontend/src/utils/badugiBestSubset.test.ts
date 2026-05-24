import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { badugiBestSubsetIndices } from './badugiBestSubset';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('badugiBestSubsetIndices', () => {
  it('returns all four indices when the hand is already a Badugi', () => {
    const hand: Card[] = [c('SPADE', 1), c('HEART', 4), c('DIAMOND', 7), c('CLOVER', 11)];
    expect(badugiBestSubsetIndices(hand)).toEqual([0, 1, 2, 3]);
  });

  it('drops one duplicate suit to leave the lowest 3-card Badugi', () => {
    // Two spades — keep the lower one.
    const hand: Card[] = [c('SPADE', 1), c('SPADE', 9), c('HEART', 3), c('DIAMOND', 6)];
    expect(badugiBestSubsetIndices(hand).sort()).toEqual([0, 2, 3]);
  });

  it('drops a duplicate rank', () => {
    const hand: Card[] = [c('SPADE', 5), c('HEART', 5), c('DIAMOND', 7), c('CLOVER', 10)];
    expect(badugiBestSubsetIndices(hand).sort()).toEqual([0, 2, 3]);
  });

  it('returns empty for empty hand', () => {
    expect(badugiBestSubsetIndices([])).toEqual([]);
  });
});
