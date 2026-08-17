import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { isGoalTopPlayableToFoundation, isSpiteAndMaliceWild } from './spiteAndMaliceUtils';

const card = (value: number, design: Card['design'] = 'SPADE'): Card => ({ design, value });

describe('isGoalTopPlayableToFoundation', () => {
  it('returns false when goalTop is undefined', () => {
    expect(isGoalTopPlayableToFoundation(undefined, [0, 0, 0, 0])).toBe(false);
  });

  it('plays an A onto an empty (top=0) foundation', () => {
    expect(isGoalTopPlayableToFoundation(card(1), [0, 0, 0, 0])).toBe(true);
  });

  it('plays a 5 only when a foundation top is 4', () => {
    expect(isGoalTopPlayableToFoundation(card(5), [3, 4, 7, 0])).toBe(true);
    expect(isGoalTopPlayableToFoundation(card(5), [3, 6, 7, 0])).toBe(false);
  });

  it('K is wild: playable on any foundation not yet at Q', () => {
    expect(isGoalTopPlayableToFoundation(card(13), [0, 0, 0, 0])).toBe(true);
    expect(isGoalTopPlayableToFoundation(card(13), [5, 11, 7, 3])).toBe(true);
  });

  it('K is not playable when every foundation is already at Q (12)', () => {
    expect(isGoalTopPlayableToFoundation(card(13), [12, 12, 12, 12])).toBe(false);
  });

  it('non-K cannot exceed +1 of any foundation top', () => {
    expect(isGoalTopPlayableToFoundation(card(7), [4, 4, 4, 4])).toBe(false);
  });
});

// #5560: K がワイルドという規則が表示に一切出ていなかった。判定はユーティリティ側に
// 置いて、画面がもう一度 13 を書かないようにする。
describe('isSpiteAndMaliceWild', () => {
  it('is true only for the King', () => {
    expect(isSpiteAndMaliceWild({ value: 13 })).toBe(true);
    expect(isSpiteAndMaliceWild({ value: 12 })).toBe(false);
    expect(isSpiteAndMaliceWild({ value: 1 })).toBe(false);
  });

  it('is false for a missing card', () => {
    expect(isSpiteAndMaliceWild(null)).toBe(false);
    expect(isSpiteAndMaliceWild(undefined)).toBe(false);
  });
});
