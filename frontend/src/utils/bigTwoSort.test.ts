import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  bigTwoCardStrength,
  bigTwoPlayTypeKey,
  bigTwoValueStrength,
  classifyBigTwoPlay,
  sortedBigTwoHand,
} from './bigTwoSort';

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

describe('classifyBigTwoPlay', () => {
  it('classifies a single card', () => {
    expect(classifyBigTwoPlay([c('SPADE', 5)])).toBe(1);
  });

  it('classifies a matching pair, rejects a mismatched pair', () => {
    expect(classifyBigTwoPlay([c('SPADE', 7), c('HEART', 7)])).toBe(2);
    expect(classifyBigTwoPlay([c('SPADE', 7), c('HEART', 8)])).toBe(0);
  });

  it('classifies a triple, rejects a mismatched triple', () => {
    expect(classifyBigTwoPlay([c('SPADE', 9), c('HEART', 9), c('CLOVER', 9)])).toBe(3);
    expect(classifyBigTwoPlay([c('SPADE', 9), c('HEART', 9), c('CLOVER', 10)])).toBe(0);
  });

  it('classifies a straight (mixed suits, consecutive)', () => {
    expect(classifyBigTwoPlay([c('SPADE', 3), c('HEART', 4), c('CLOVER', 5), c('DIAMOND', 6), c('SPADE', 7)])).toBe(4);
  });

  it('classifies the Ace-high 10-J-Q-K-A straight', () => {
    expect(classifyBigTwoPlay([c('SPADE', 10), c('HEART', 11), c('CLOVER', 12), c('DIAMOND', 13), c('SPADE', 1)])).toBe(
      4,
    );
  });

  it('rejects a run containing a 2 as a straight', () => {
    expect(classifyBigTwoPlay([c('SPADE', 1), c('HEART', 2), c('CLOVER', 3), c('DIAMOND', 4), c('SPADE', 5)])).toBe(0);
  });

  it('classifies a flush (same suit, not consecutive)', () => {
    expect(classifyBigTwoPlay([c('SPADE', 3), c('SPADE', 5), c('SPADE', 7), c('SPADE', 9), c('SPADE', 11)])).toBe(5);
  });

  it('classifies a full house', () => {
    expect(classifyBigTwoPlay([c('SPADE', 8), c('HEART', 8), c('CLOVER', 8), c('DIAMOND', 4), c('SPADE', 4)])).toBe(6);
  });

  it('classifies four of a kind (plus kicker)', () => {
    expect(classifyBigTwoPlay([c('SPADE', 6), c('HEART', 6), c('CLOVER', 6), c('DIAMOND', 6), c('SPADE', 9)])).toBe(7);
  });

  it('classifies a straight flush', () => {
    expect(classifyBigTwoPlay([c('SPADE', 3), c('SPADE', 4), c('SPADE', 5), c('SPADE', 6), c('SPADE', 7)])).toBe(8);
  });

  it('returns 0 for empty, 4-card, and 6-card selections', () => {
    expect(classifyBigTwoPlay([])).toBe(0);
    expect(classifyBigTwoPlay([c('SPADE', 3), c('HEART', 3), c('CLOVER', 3), c('DIAMOND', 5)])).toBe(0);
    expect(
      classifyBigTwoPlay([c('SPADE', 3), c('HEART', 4), c('CLOVER', 5), c('DIAMOND', 6), c('SPADE', 7), c('HEART', 8)]),
    ).toBe(0);
  });

  it('returns 0 for a 5-card selection that is neither straight, flush, nor a set', () => {
    expect(classifyBigTwoPlay([c('SPADE', 3), c('HEART', 5), c('CLOVER', 7), c('DIAMOND', 9), c('SPADE', 11)])).toBe(0);
  });
});
