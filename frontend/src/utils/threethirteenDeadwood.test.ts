import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import {
  bestThreeThirteenDeadwoodValue,
  bestThreeThirteenDiscardValue,
  calcThreeThirteenDeadwoodValue,
  THREETHIRTEEN_WILD_DEADWOOD_VALUE,
  threeThirteenCardValue,
  threeThirteenIsWild,
} from './threethirteenDeadwood';

const c = (design: CardDesign, value: number): Card => ({ design, value });

describe('threethirteenDeadwood', () => {
  it('scores single card values (A=1, face, J/Q/K=10)', () => {
    expect(threeThirteenCardValue(c('SPADE', 1))).toBe(1);
    expect(threeThirteenCardValue(c('SPADE', 7))).toBe(7);
    expect(threeThirteenCardValue(c('SPADE', 10))).toBe(10);
    expect(threeThirteenCardValue(c('SPADE', 13))).toBe(10);
  });

  it('identifies the wild rank', () => {
    expect(threeThirteenIsWild(c('SPADE', 4), 4)).toBe(true);
    expect(threeThirteenIsWild(c('SPADE', 5), 4)).toBe(false);
  });

  it('counts a leftover wild as the wild penalty', () => {
    // ♥3 (=3) + a wild (rank 4) left as deadwood → 3 + 20.
    const total = calcThreeThirteenDeadwoodValue([c('HEART', 3), c('DIAMOND', 4)], 4);
    expect(total).toBe(3 + THREETHIRTEEN_WILD_DEADWOOD_VALUE);
  });

  it('returns 0 deadwood for a natural set of three', () => {
    const hand = [c('SPADE', 8), c('HEART', 8), c('CLOVER', 8)];
    expect(bestThreeThirteenDeadwoodValue(hand, 4)).toBe(0);
  });

  it('completes a run with a wild card (deadwood 0)', () => {
    // ♠5, ♠7 plus a wild (rank 4) filling the 6 → run 5-6-7, deadwood 0.
    const hand = [c('SPADE', 5), c('SPADE', 7), c('HEART', 4)];
    expect(bestThreeThirteenDeadwoodValue(hand, 4)).toBe(0);
  });

  it('completes a set with a wild card (deadwood 0)', () => {
    // ♠9, ♥9 plus a wild (rank 2) → set of 9s, deadwood 0.
    const hand = [c('SPADE', 9), c('HEART', 9), c('CLOVER', 2)];
    expect(bestThreeThirteenDeadwoodValue(hand, 2)).toBe(0);
  });

  it('leaves the non-melding card as deadwood', () => {
    // ♠5,♠6,♠7 form a run; ♥K (10) is unmatched → deadwood 10.
    const hand = [c('SPADE', 5), c('SPADE', 6), c('SPADE', 7), c('HEART', 13)];
    expect(bestThreeThirteenDeadwoodValue(hand, 2)).toBe(10);
  });

  it('treats Ace as high in a Q-K-A run', () => {
    const hand = [c('SPADE', 12), c('SPADE', 13), c('SPADE', 1)];
    expect(bestThreeThirteenDeadwoodValue(hand, 2)).toBe(0);
  });

  it('finds the best single discard', () => {
    // Discarding ♥K leaves the ♠5-6-7 run → 0; any other discard is worse.
    const hand = [c('SPADE', 5), c('SPADE', 6), c('SPADE', 7), c('HEART', 13)];
    expect(bestThreeThirteenDiscardValue(hand, 2)).toBe(0);
  });

  it('returns 0 for empty hands', () => {
    expect(bestThreeThirteenDeadwoodValue([], 4)).toBe(0);
    expect(bestThreeThirteenDiscardValue([], 4)).toBe(0);
  });
});
