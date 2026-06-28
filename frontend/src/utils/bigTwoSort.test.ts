import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { bigTwoCardStrength, bigTwoPlayTypeKey, bigTwoValueStrength, sortedBigTwoHand } from './bigTwoSort';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('bigTwoValueStrength', () => {
  it('ranks 2 highest, then A, then 3..K ascending', () => {
    expect(bigTwoValueStrength(2)).toBe(12);
    expect(bigTwoValueStrength(1)).toBe(11);
    expect(bigTwoValueStrength(3)).toBe(0);
    expect(bigTwoValueStrength(13)).toBe(10);
  });
});

describe('bigTwoCardStrength', () => {
  it('orders by value first then suit (♦<♣<♥<♠)', () => {
    expect(bigTwoCardStrength(c('SPADE', 2))).toBeGreaterThan(bigTwoCardStrength(c('DIAMOND', 2)));
    expect(bigTwoCardStrength(c('DIAMOND', 2))).toBeGreaterThan(bigTwoCardStrength(c('SPADE', 1)));
  });
});

describe('sortedBigTwoHand', () => {
  const hand = [c('SPADE', 3), c('DIAMOND', 2), c('HEART', 1), c('CLOVER', 3)];

  it('keeps each card paired with its original index', () => {
    const out = sortedBigTwoHand(hand, 'strength');
    for (const { card, index } of out) {
      expect(hand[index]).toBe(card);
    }
  });

  it('strength mode puts the 2 last and the lowest 3 first', () => {
    const out = sortedBigTwoHand(hand, 'strength');
    expect(out[0].card.value).toBe(3);
    expect(out[out.length - 1].card).toEqual(c('DIAMOND', 2));
  });

  it('suit mode groups by suit ♦♣♥♠', () => {
    const out = sortedBigTwoHand(hand, 'suit');
    expect(out.map((o) => o.card.design)).toEqual(['DIAMOND', 'CLOVER', 'HEART', 'SPADE']);
  });

  it('number mode uses natural rank (A first, 2 second)', () => {
    const out = sortedBigTwoHand(hand, 'number');
    expect(out[0].card.value).toBe(1); // Ace
    expect(out[1].card.value).toBe(2);
  });
});

describe('bigTwoPlayTypeKey', () => {
  it('maps every play type 1-8 to its key', () => {
    expect([1, 2, 3, 4, 5, 6, 7, 8].map(bigTwoPlayTypeKey)).toEqual([
      'single',
      'pair',
      'triple',
      'straight',
      'flush',
      'fullHouse',
      'fourOfAKind',
      'straightFlush',
    ]);
  });

  it('returns null for invalid/empty (0)', () => {
    expect(bigTwoPlayTypeKey(0)).toBeNull();
    expect(bigTwoPlayTypeKey(99)).toBeNull();
  });
});
