import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { kalookiCardValue, kalookiMeldValue, kalookiOpeningPoints } from './kalookiScore';

const c = (design: CardDesign, value: number): Card => ({ design, value });

describe('kalookiCardValue', () => {
  it('scores an ace as 15, not 1', () => {
    expect(kalookiCardValue(c('SPADE', 1))).toBe(15);
  });

  it('scores a joker as 15', () => {
    expect(kalookiCardValue(c('JOKER', 0))).toBe(15);
  });

  it('scores ten and above as 10', () => {
    expect([10, 11, 12, 13].map((v) => kalookiCardValue(c('HEART', v)))).toEqual([10, 10, 10, 10]);
  });

  it('scores the rest at face value', () => {
    expect([2, 5, 9].map((v) => kalookiCardValue(c('HEART', v)))).toEqual([2, 5, 9]);
  });
});

describe('kalookiMeldValue', () => {
  it('sums a joker-free meld', () => {
    expect(kalookiMeldValue([c('SPADE', 5), c('HEART', 5), c('CLOVER', 5)])).toBe(15);
  });

  it('applies the 1.5x joker bonus, floored', () => {
    // 5 + 5 + 15 = 25 → 37.5 → 37, matching the domain's integer arithmetic.
    expect(kalookiMeldValue([c('SPADE', 5), c('HEART', 5), c('JOKER', 0)])).toBe(37);
  });
});

describe('kalookiOpeningPoints', () => {
  it('is zero with nothing staged', () => {
    expect(kalookiOpeningPoints([])).toBe(0);
  });

  it('adds the melds up, applying the joker bonus per meld rather than overall', () => {
    const plain = [c('SPADE', 10), c('HEART', 10), c('CLOVER', 10)]; // 30
    const withJoker = [c('SPADE', 5), c('HEART', 5), c('JOKER', 0)]; // 37
    expect(kalookiOpeningPoints([plain, withJoker])).toBe(67);
    // Summing the cards first and then applying one bonus would give 82 — the
    // mistake a player doing this by eye would make.
    expect(kalookiOpeningPoints([plain, withJoker])).not.toBe(82);
  });
});
