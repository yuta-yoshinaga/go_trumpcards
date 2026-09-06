import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { omiLegalPlayIndices } from './omiLegalPlay';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('omiLegalPlayIndices', () => {
  const hand = [c('SPADE', 1), c('HEART', 10), c('DIAMOND', 9)];

  it('marks every card legal when leading (no lead card)', () => {
    expect(omiLegalPlayIndices(hand, null, 1)).toEqual([0, 1, 2]);
  });

  it('restricts to the led suit when the player can follow', () => {
    // Lead HEART → only HEART(idx 1) follows.
    expect(omiLegalPlayIndices(hand, c('HEART', 1), 1)).toEqual([1]);
  });

  it('marks every card legal when the player cannot follow', () => {
    // Lead CLOVER → hand has no clover, all legal.
    expect(omiLegalPlayIndices(hand, c('CLOVER', 1), 4)).toEqual([0, 1, 2]);
  });

  // In Omi there are no bowers — a Jack of a same-color suit is NOT treated
  // as trump for follow-suit purposes (no left bower mechanic).
  it('treats a Jack of same-color suit as its printed suit, NOT as trump', () => {
    // trump SPADE(1); in Euchre, Jack of CLOVER would be left bower (counts as trump).
    // In Omi it is just a Clover card.
    const h = [c('CLOVER', 11), c('HEART', 1)];
    // Lead SPADE → neither card is spade, so both are legal.
    expect(omiLegalPlayIndices(h, c('SPADE', 14), 1)).toEqual([0, 1]);
  });

  it('marks every card legal when leading (undefined lead card)', () => {
    expect(omiLegalPlayIndices(hand, undefined, 1)).toEqual([0, 1, 2]);
  });
});
