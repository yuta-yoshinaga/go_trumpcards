import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { shamrocksHasLegalMove, shamrocksMovableFans } from './shamrocksLegalMove';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const noFoundation = (): Card[][] => [[], [], [], []];

// The page borrowed labelleLucieMovableFans, which encodes the SOURCE game's
// rule. Each case below is one Shamrocks rule that borrowing got wrong.
describe('shamrocksMovableFans', () => {
  it('accepts a different suit one rank away (La Belle Lucie needs the same suit)', () => {
    const fans = [[card('HEART', 6)], [card('SPADE', 7)]];
    expect(shamrocksMovableFans(fans, noFoundation())).toEqual(new Set([0, 1]));
  });

  it('accepts one rank HIGHER, not just lower', () => {
    const fans = [[card('HEART', 8)], [card('SPADE', 7)]];
    expect(shamrocksMovableFans(fans, noFoundation()).has(0)).toBe(true);
  });

  it('treats an empty fan as a destination', () => {
    // Nothing stacks, but the empty fan takes anything.
    const fans = [[card('HEART', 6)], []];
    expect(shamrocksHasLegalMove(fans, noFoundation())).toBe(true);
  });

  it('refuses a full fan even when the rank fits', () => {
    const fans = [
      [card('HEART', 11), card('HEART', 13), card('HEART', 6)],
      [card('SPADE', 2), card('SPADE', 4), card('SPADE', 7)],
    ];
    // 6 and 7 are adjacent, but both fans hold three, so neither can receive.
    expect(shamrocksHasLegalMove(fans, noFoundation())).toBe(false);
  });

  it('still requires adjacency', () => {
    const fans = [
      [card('HEART', 11), card('HEART', 13), card('HEART', 3)],
      [card('SPADE', 2), card('SPADE', 4), card('SPADE', 7)],
    ];
    expect(shamrocksHasLegalMove(fans, noFoundation())).toBe(false);
  });

  it('finds a foundation move', () => {
    const fans = [[card('SPADE', 1)]];
    expect(shamrocksMovableFans(fans, noFoundation()).has(0)).toBe(true);
  });
});
