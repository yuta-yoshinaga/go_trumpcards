import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { holdemBestFive } from './holdemBestFive';

const c = (design: Card['design'], value: number): Card => ({ design, value });

/**
 * Calls {@link holdemBestFive} and narrows away the `null` return.
 *
 * Every case below except the explicit "fewer than 5 cards" one expects a
 * hand, so the alternative is a `!` on each call — which Biome flags, and
 * whose auto-fix (`?.`) would only push the `undefined` into the assertions
 * and fail tsc. Throwing here keeps the call sites clean and turns a
 * regression that starts returning null into a named failure rather than a
 * "cannot read property of undefined".
 */
const bestFive = (cards: Card[]): number[] => {
  const picked = holdemBestFive(cards);
  if (picked === null) throw new Error('holdemBestFive unexpectedly returned null');
  return picked;
};

describe('holdemBestFive', () => {
  it('returns null for fewer than 5 cards', () => {
    expect(holdemBestFive([c('SPADE', 2), c('CLOVER', 3)])).toBeNull();
  });

  it('picks a four-of-a-kind over an unrelated flush', () => {
    const cards = [
      c('SPADE', 9),
      c('CLOVER', 9),
      c('HEART', 9),
      c('DIAMOND', 9),
      c('SPADE', 5),
      c('SPADE', 7),
      c('SPADE', 3),
    ];
    const picked = bestFive(cards);
    const values = picked.map((i) => cards[i].value).sort((a, b) => a - b);
    // Quad 9s + highest available kicker (7).
    expect(values).toEqual([7, 9, 9, 9, 9]);
  });

  it('prefers a flush over two pair', () => {
    const cards = [
      c('SPADE', 2),
      c('SPADE', 4),
      c('SPADE', 6),
      c('SPADE', 8),
      c('SPADE', 10),
      c('HEART', 4),
      c('HEART', 6),
    ];
    const picked = bestFive(cards).map((i) => cards[i]);
    expect(picked.every((card) => card.design === 'SPADE')).toBe(true);
  });

  it('recognizes the wheel straight (A-2-3-4-5)', () => {
    const cards = [
      c('SPADE', 1),
      c('CLOVER', 2),
      c('HEART', 3),
      c('DIAMOND', 4),
      c('SPADE', 5),
      c('HEART', 9),
      c('CLOVER', 10),
    ];
    const picked = bestFive(cards)
      .map((i) => cards[i].value)
      .sort((a, b) => a - b);
    expect(picked).toEqual([1, 2, 3, 4, 5]);
  });

  it('picks the higher kicker on one-pair ties', () => {
    const cards = [
      c('SPADE', 10),
      c('CLOVER', 10),
      c('HEART', 13), // K
      c('DIAMOND', 9),
      c('SPADE', 4),
      c('CLOVER', 7),
      c('HEART', 2),
    ];
    const picked = bestFive(cards)
      .map((i) => cards[i].value)
      .sort((a, b) => a - b);
    // Pair of 10s plus K, 9, 7 kickers.
    expect(picked).toEqual([7, 9, 10, 10, 13]);
  });

  it('picks 5 community cards when those are better than including hole cards', () => {
    // Royal flush on the board.
    const cards = [
      c('HEART', 2), // hole
      c('SPADE', 3), // hole
      c('HEART', 10),
      c('HEART', 11),
      c('HEART', 12),
      c('HEART', 13),
      c('HEART', 1), // Ace high
    ];
    const picked = bestFive(cards).map((i) => cards[i]);
    expect(picked.every((card) => card.design === 'HEART')).toBe(true);
  });
});
