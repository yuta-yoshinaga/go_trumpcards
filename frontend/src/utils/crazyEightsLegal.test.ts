import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { isCrazyEightsLegalPlay } from './crazyEightsLegal';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('isCrazyEightsLegalPlay', () => {
  const top = c('HEART', 5);

  it('always allows an 8 (wild)', () => {
    expect(isCrazyEightsLegalPlay(c('SPADE', 8), top, 0)).toBe(true);
  });

  it('allows a matching suit or rank against the discard top', () => {
    expect(isCrazyEightsLegalPlay(c('HEART', 10), top, 0)).toBe(true); // same suit
    expect(isCrazyEightsLegalPlay(c('SPADE', 5), top, 0)).toBe(true); // same rank
  });

  it('rejects a card matching neither suit nor rank', () => {
    expect(isCrazyEightsLegalPlay(c('SPADE', 10), top, 0)).toBe(false);
  });

  it('uses the chosen suit when an 8 set one', () => {
    expect(isCrazyEightsLegalPlay(c('CLOVER', 3), top, 2)).toBe(true); // clover = suit 2
    expect(isCrazyEightsLegalPlay(c('HEART', 3), top, 2)).toBe(false); // heart != chosen clover
  });

  it('allows any card on the opening play (no discard top)', () => {
    expect(isCrazyEightsLegalPlay(c('SPADE', 10), null, 0)).toBe(true);
  });
});
