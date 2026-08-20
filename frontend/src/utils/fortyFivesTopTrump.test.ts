import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { fortyFivesTopTrumpIndices, isFortyFivesTopTrump } from './fortyFivesTopTrump';

const card = (design: Card['design'], value: number): Card => ({ design, value });

// domain の CardDesign* と同じ番号。0 は切り札未決定。
const NO_TRUMP = 0;
const SPADE = 1;
const CLOVER = 2;
const HEART = 3;
const DIAMOND = 4;

describe('isFortyFivesTopTrump', () => {
  it('counts the five and the jack of the trump suit', () => {
    expect(isFortyFivesTopTrump(card('SPADE', 5), SPADE)).toBe(true);
    expect(isFortyFivesTopTrump(card('SPADE', 11), SPADE)).toBe(true);
  });

  // ここが Forty-Fives 固有の分かりにくさ: ♥A は切り札スートに関係なく常に最上位。
  it('counts the ace of hearts whatever the trump suit is', () => {
    for (const trump of [SPADE, CLOVER, HEART, DIAMOND]) {
      expect(isFortyFivesTopTrump(card('HEART', 1), trump)).toBe(true);
    }
  });

  it('does not count the trump ace, nor a five or jack of another suit', () => {
    expect(isFortyFivesTopTrump(card('SPADE', 1), SPADE)).toBe(false);
    expect(isFortyFivesTopTrump(card('CLOVER', 5), SPADE)).toBe(false);
    expect(isFortyFivesTopTrump(card('CLOVER', 11), SPADE)).toBe(false);
  });

  // 切り札が決まる前 (trumpSuit === 0) は 5/J を最上位と呼べない。♥A だけは常に真。
  it('marks only the ace of hearts before a trump suit is set', () => {
    expect(isFortyFivesTopTrump(card('SPADE', 5), NO_TRUMP)).toBe(false);
    expect(isFortyFivesTopTrump(card('HEART', 1), NO_TRUMP)).toBe(true);
  });
});

describe('fortyFivesTopTrumpIndices', () => {
  it('returns the positions in hand order', () => {
    const hand = [card('SPADE', 5), card('CLOVER', 9), card('HEART', 1), card('SPADE', 1), card('SPADE', 11)];
    expect(fortyFivesTopTrumpIndices(hand, SPADE)).toEqual([0, 2, 4]);
  });

  it('returns an empty array for a hand with none', () => {
    expect(fortyFivesTopTrumpIndices([card('CLOVER', 9), card('DIAMOND', 3)], SPADE)).toEqual([]);
  });
});
