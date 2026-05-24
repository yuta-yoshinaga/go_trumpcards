import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { paiGowFoulCheck } from './paiGowFoul';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('paiGowFoulCheck', () => {
  it('returns no foul when low is high card and high beats it', () => {
    // Hand: S2,S3,S5,S6,S7,H10,DK; low = first two (S2,S3 → top 3); high contains DK=13
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
    // Low = [S K, H K] -> pair of K; High = [D2, S3, C4, D6, H8] -> high card → foul
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

  it('does not flag foul when high contains a pair', () => {
    // Low = [S K, H K]; High = [D2, S3, C4, D6, H6] (pair of 6) → high beats low only if rank-wise high pair beats low pair?
    // High-rank wins because high>=pair short-circuits in our check (we only forbid low > high card).
    const cards = [
      c('SPADE', 13),
      c('HEART', 13),
      c('DIAMOND', 2),
      c('SPADE', 3),
      c('CLOVER', 4),
      c('DIAMOND', 6),
      c('HEART', 6),
    ];
    expect(paiGowFoulCheck(cards, [0, 1]).isFoul).toBe(false);
  });

  it('flags foul when both are high cards but low top > high top', () => {
    // Low = [D A, S K] tops=14,13; High = [H2,C3,D4,S5,H7] tops=7 → foul.
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

  it('treats a 5-card flush as beating any low', () => {
    // High = all spades (5 spades); Low = [H K, H K]
    const cards = [
      c('HEART', 13),
      c('HEART', 13),
      c('SPADE', 2),
      c('SPADE', 4),
      c('SPADE', 7),
      c('SPADE', 9),
      c('SPADE', 11),
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
