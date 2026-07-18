import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { cinchCardPoints, estimateCinchBidStrength } from './cinchBidStrength';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('cinchCardPoints', () => {
  it('scores the trump A/K/10/J as 1 point each', () => {
    expect(cinchCardPoints(card('HEART', 1), 3)).toBe(1); // High (A)
    expect(cinchCardPoints(card('HEART', 13), 3)).toBe(1); // King
    expect(cinchCardPoints(card('HEART', 10), 3)).toBe(1); // Ten (Game)
    expect(cinchCardPoints(card('HEART', 11), 3)).toBe(1); // Jack
  });

  it('scores the Right Pedro (5 of trump) as 5 points', () => {
    expect(cinchCardPoints(card('HEART', 5), 3)).toBe(5);
  });

  it('scores the Left Pedro (5 of same-color off-suit) as 5 points', () => {
    // Trump hearts -> same-color diamonds; the 5♦ is the Left Pedro.
    expect(cinchCardPoints(card('DIAMOND', 5), 3)).toBe(5);
    // Trump spades -> same-color clubs; the 5♣ is the Left Pedro.
    expect(cinchCardPoints(card('CLOVER', 5), 1)).toBe(5);
  });

  it('scores non-point trump cards and off-suit cards as 0', () => {
    expect(cinchCardPoints(card('HEART', 9), 3)).toBe(0); // trump but not a point card
    expect(cinchCardPoints(card('SPADE', 1), 3)).toBe(0); // off-suit ace
    expect(cinchCardPoints(card('JOKER', 0), 3)).toBe(0); // unknown design
  });
});

describe('estimateCinchBidStrength', () => {
  it('picks the strongest suit and brackets the range', () => {
    // Strong in hearts: A♥, K♥, 5♥ (Right Pedro) = 1+1+5 = 7, and 5♦ is the
    // Left Pedro for hearts (+5) => hearts total 12. Diamonds trump: 5♦ Right
    // Pedro (5) + 5♥ Left Pedro (5) = 10.
    const hand = [card('HEART', 1), card('HEART', 13), card('HEART', 5), card('DIAMOND', 5), card('SPADE', 2)];
    const s = estimateCinchBidStrength(hand);
    expect(s.bestSuit).toBe(3); // hearts
    expect(s.pointsBySuit[3]).toBe(12);
    expect(s.pointsBySuit[4]).toBe(10); // diamonds
    expect(s.maxPoints).toBe(12);
    expect(s.minPoints).toBe(0); // clubs holds no point cards here
  });

  it('returns all-zero strength for a hand with no point cards', () => {
    const hand = [card('SPADE', 2), card('CLOVER', 3), card('HEART', 8), card('DIAMOND', 9)];
    const s = estimateCinchBidStrength(hand);
    expect(s.maxPoints).toBe(0);
    expect(s.minPoints).toBe(0);
    expect(s.bestSuit).toBe(1); // ties resolve to the lowest suit index
  });
});
