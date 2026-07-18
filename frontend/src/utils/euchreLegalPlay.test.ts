import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { euchreEffectiveSuit, euchreLegalPlayIndices } from './euchreLegalPlay';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('euchreEffectiveSuit', () => {
  it('returns the printed suit for a non-bower card', () => {
    expect(euchreEffectiveSuit(c('HEART', 1), 1)).toBe(3);
  });

  it('treats the left bower (same-color Jack) as the trump suit', () => {
    // trump = SPADE(1); same-color = CLOVER(2). Jack of clubs is the left bower.
    expect(euchreEffectiveSuit(c('CLOVER', 11), 1)).toBe(1);
  });

  it('keeps the right bower on its printed (trump) suit', () => {
    expect(euchreEffectiveSuit(c('SPADE', 11), 1)).toBe(1);
  });

  it('does not treat an off-color Jack as trump', () => {
    expect(euchreEffectiveSuit(c('HEART', 11), 1)).toBe(3);
  });

  it('uses the printed suit when no trump is set', () => {
    expect(euchreEffectiveSuit(c('CLOVER', 11), 0)).toBe(2);
  });
});

describe('euchreLegalPlayIndices', () => {
  const hand = [c('SPADE', 1), c('HEART', 11), c('DIAMOND', 9)];

  it('marks every card legal when leading (no lead card)', () => {
    expect(euchreLegalPlayIndices(hand, null, 1)).toEqual([0, 1, 2]);
  });

  it('restricts to the led suit when the player can follow', () => {
    // Lead HEART, trump SPADE → only HEART(idx 1) follows.
    expect(euchreLegalPlayIndices(hand, c('HEART', 1), 1)).toEqual([1]);
  });

  it('marks every card legal when the player cannot follow', () => {
    // Lead CLOVER, trump DIAMOND → hand has no clover, all legal.
    expect(euchreLegalPlayIndices(hand, c('CLOVER', 1), 4)).toEqual([0, 1, 2]);
  });

  it('treats the left bower as trump for follow-suit', () => {
    // trump SPADE(1); left bower = Jack of CLOVER(2). Lead a plain spade.
    const h = [c('CLOVER', 11), c('HEART', 1)];
    // Lead SPADE → left bower (idx 0) counts as trump/spade and must follow.
    expect(euchreLegalPlayIndices(h, c('SPADE', 14), 1)).toEqual([0]);
  });
});
