import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { tichuBombIndices } from './tichuBomb';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('tichuBombIndices', () => {
  it('flags a four-of-a-kind', () => {
    const hand = [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7), card('DIAMOND', 7), card('SPADE', 9)];
    expect([...tichuBombIndices(hand)].sort((a, b) => a - b)).toEqual([0, 1, 2, 3]);
  });

  it('flags a five-card straight flush', () => {
    const hand = [
      card('SPADE', 5),
      card('SPADE', 6),
      card('SPADE', 7),
      card('SPADE', 8),
      card('SPADE', 9),
      card('HEART', 2),
    ];
    expect([...tichuBombIndices(hand)].sort((a, b) => a - b)).toEqual([0, 1, 2, 3, 4]);
  });

  it('treats the Ace as high (10-J-Q-K-A is a straight flush, A-2-3-4-5 is not)', () => {
    const highRun = [
      card('HEART', 10),
      card('HEART', 11),
      card('HEART', 12),
      card('HEART', 13),
      card('HEART', 1), // Ace
    ];
    expect(tichuBombIndices(highRun).size).toBe(5);

    const wheel = [card('HEART', 1), card('HEART', 2), card('HEART', 3), card('HEART', 4), card('HEART', 5)];
    expect(tichuBombIndices(wheel).size).toBe(0); // Ace is high only — no wheel bomb
  });

  it('does not flag a four-card straight or a mixed-suit run', () => {
    const shortRun = [card('SPADE', 5), card('SPADE', 6), card('SPADE', 7), card('SPADE', 8)];
    expect(tichuBombIndices(shortRun).size).toBe(0);
    const mixed = [card('SPADE', 5), card('HEART', 6), card('SPADE', 7), card('SPADE', 8), card('SPADE', 9)];
    expect(tichuBombIndices(mixed).size).toBe(0);
  });

  it('ignores special (JOKER) cards', () => {
    // Three sevens + two JOKER specials must not be mistaken for a bomb.
    const hand = [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7), card('JOKER', 4), card('JOKER', 3)];
    expect(tichuBombIndices(hand).size).toBe(0);
  });

  it('returns an empty set for a bombless hand', () => {
    const hand = [card('SPADE', 2), card('HEART', 5), card('CLOVER', 9), card('DIAMOND', 13)];
    expect(tichuBombIndices(hand).size).toBe(0);
  });
});
