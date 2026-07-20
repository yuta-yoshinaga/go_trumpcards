import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { paiGowAutoSplit, paiGowFoulCheck } from './paiGowFoul';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('paiGowFoulCheck', () => {
  it('returns no foul when low is high card and high beats it', () => {
    // Low = [S2, S3] top=3; High = [S5,S6,S7,H10,DK] top=13 → not foul
    const cards = [
      c('SPADE', 2),
      c('SPADE', 3),
      c('SPADE', 5),
      c('SPADE', 6),
      c('SPADE', 7),
      c('HEART', 10),
      c('DIAMOND', 13),
    ];
    expect(paiGowFoulCheck(cards, [0, 1])).toEqual({ isFoul: false });
  });

  it('flags foul when low is a pair and high is just high card', () => {
    // Low = [S K, H K] → pair of K; High = [D2, S3, C4, D6, H8] → high card → foul
    const cards = [
      c('SPADE', 13),
      c('HEART', 13),
      c('DIAMOND', 2),
      c('SPADE', 3),
      c('CLOVER', 4),
      c('DIAMOND', 6),
      c('HEART', 8),
    ];
    expect(paiGowFoulCheck(cards, [0, 1])).toEqual({ isFoul: true });
  });

  it('flags foul when low pair outranks high pair', () => {
    // Low = [S K, H K] (Pair of 13); High = [D2, S3, C4, D6, H6] (Pair of 6) → foul
    const cards = [
      c('SPADE', 13),
      c('HEART', 13),
      c('DIAMOND', 2),
      c('SPADE', 3),
      c('CLOVER', 4),
      c('DIAMOND', 6),
      c('HEART', 6),
    ];
    expect(paiGowFoulCheck(cards, [0, 1]).isFoul).toBe(true);
  });

  it('does not flag foul when high pair outranks low pair', () => {
    // Low = [S2, H2] (Pair of 2s); High = [D6, S6, C4, D7, H8] (Pair of 6s) → not foul
    const cards = [
      c('SPADE', 2),
      c('HEART', 2),
      c('DIAMOND', 6),
      c('SPADE', 6),
      c('CLOVER', 4),
      c('DIAMOND', 7),
      c('HEART', 8),
    ];
    expect(paiGowFoulCheck(cards, [0, 1]).isFoul).toBe(false);
  });

  it('flags foul when both are high cards but low top > high top', () => {
    // Low = [D A, S K] tops=14,13; High = [H2,C3,D4,S5,H7] tops=7 → foul
    const cards = [
      c('DIAMOND', 1),
      c('SPADE', 13),
      c('HEART', 2),
      c('CLOVER', 3),
      c('DIAMOND', 4),
      c('SPADE', 5),
      c('HEART', 7),
    ];
    expect(paiGowFoulCheck(cards, [0, 1]).isFoul).toBe(true);
  });

  it('flags foul when top cards tie but low second card outranks high second card', () => {
    // Low = [D K, S Q] vals=13,12; High = [H K, C3, D4, S5, H7] vals=13,7,5,4,3 → foul (12 > 7)
    const cards = [
      c('DIAMOND', 13),
      c('SPADE', 12),
      c('HEART', 13),
      c('CLOVER', 3),
      c('DIAMOND', 4),
      c('SPADE', 5),
      c('HEART', 7),
    ];
    expect(paiGowFoulCheck(cards, [0, 1]).isFoul).toBe(true);
  });

  it('does not flag foul when top cards tie and high second card outranks low second card', () => {
    // Low = [D K, S3] vals=13,3; High = [H K, C Q, D4, S5, H7] vals=13,12,7,5,4 → not foul (12 > 3)
    const cards = [
      c('DIAMOND', 13),
      c('SPADE', 3),
      c('HEART', 13),
      c('CLOVER', 12),
      c('DIAMOND', 4),
      c('SPADE', 5),
      c('HEART', 7),
    ];
    expect(paiGowFoulCheck(cards, [0, 1]).isFoul).toBe(false);
  });

  it('treats a 5-card flush as beating any low', () => {
    // High = all spades (flush); Low = [H K, H Q] → not foul
    const cards = [
      c('HEART', 13),
      c('HEART', 12),
      c('SPADE', 2),
      c('SPADE', 4),
      c('SPADE', 7),
      c('SPADE', 9),
      c('SPADE', 11),
    ];
    expect(paiGowFoulCheck(cards, [0, 1]).isFoul).toBe(false);
  });

  it('treats a 5-card straight as beating any low', () => {
    // High = [H2,S3,D4,C5,H6] (straight); Low = [D A, S K] → not foul
    const cards = [
      c('DIAMOND', 1),
      c('SPADE', 13),
      c('HEART', 2),
      c('SPADE', 3),
      c('DIAMOND', 4),
      c('CLOVER', 5),
      c('HEART', 6),
    ];
    expect(paiGowFoulCheck(cards, [0, 1]).isFoul).toBe(false);
  });

  it('treats two pair in high hand as beating any low', () => {
    // High = [S2,H2,D3,C3,H8] (two pair); Low = [D A, S K] → not foul
    const cards = [
      c('DIAMOND', 1),
      c('SPADE', 13),
      c('SPADE', 2),
      c('HEART', 2),
      c('DIAMOND', 3),
      c('CLOVER', 3),
      c('HEART', 8),
    ];
    expect(paiGowFoulCheck(cards, [0, 1]).isFoul).toBe(false);
  });

  it('skips check when a joker is present', () => {
    const cards = [
      c('JOKER', 0),
      c('HEART', 13),
      c('SPADE', 2),
      c('DIAMOND', 3),
      c('CLOVER', 4),
      c('SPADE', 5),
      c('HEART', 7),
    ];
    expect(paiGowFoulCheck(cards, [0, 1])).toEqual({ isFoul: false });
  });

  it('returns no foul when lowIndices length is not 2', () => {
    const cards = [
      c('SPADE', 2),
      c('SPADE', 3),
      c('SPADE', 5),
      c('SPADE', 6),
      c('SPADE', 7),
      c('HEART', 10),
      c('DIAMOND', 13),
    ];
    expect(paiGowFoulCheck(cards, [0])).toEqual({ isFoul: false });
  });
});

describe('paiGowAutoSplit', () => {
  it('picks the strongest legal high-card low hand and never fouls (all singletons)', () => {
    // S10, H11, D13, C5, H7, S3, D9 — the 13 must stay high, so the best legal
    // low hand is {S10, H11} at indices [0, 1].
    const cards = [
      c('SPADE', 10),
      c('HEART', 11),
      c('DIAMOND', 13),
      c('CLOVER', 5),
      c('HEART', 7),
      c('SPADE', 3),
      c('DIAMOND', 9),
    ];
    const split = paiGowAutoSplit(cards);
    expect(split).toEqual([0, 1]);
    expect(paiGowFoulCheck(cards, split as [number, number])).toEqual({ isFoul: false });
  });

  it('prefers a legal low pair over any high-card low hand', () => {
    // Pair of 5s can sit in the low hand because the high hand holds a higher pair (aces).
    const cards = [
      c('SPADE', 5),
      c('HEART', 5),
      c('SPADE', 1),
      c('HEART', 1),
      c('DIAMOND', 2),
      c('CLOVER', 3),
      c('HEART', 4),
    ];
    const split = paiGowAutoSplit(cards);
    expect(split).toEqual([0, 1]);
    expect(paiGowFoulCheck(cards, split as [number, number])).toEqual({ isFoul: false });
  });

  it('returns null when a joker is present (foul evaluation unavailable)', () => {
    const cards = [
      c('JOKER', 0),
      c('HEART', 13),
      c('SPADE', 2),
      c('DIAMOND', 3),
      c('CLOVER', 4),
      c('SPADE', 5),
      c('HEART', 7),
    ];
    expect(paiGowAutoSplit(cards)).toBeNull();
  });

  it('returns null when the hand is not exactly 7 cards', () => {
    expect(paiGowAutoSplit([c('SPADE', 2), c('SPADE', 3)])).toBeNull();
  });
});
