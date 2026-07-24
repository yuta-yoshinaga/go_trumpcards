import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { evaluateMississippiStudMadeHand } from './mississippiStudMadeHand';
import { PokerHand } from './pokerSquaresUtils';

const card = (design: CardDesign, value: number): Card => ({ design, value });

describe('evaluateMississippiStudMadeHand', () => {
  it('returns null when fewer than two cards are known', () => {
    expect(evaluateMississippiStudMadeHand([card('SPADE', 8)])).toBeNull();
    expect(evaluateMississippiStudMadeHand([])).toBeNull();
  });

  it('returns null for a high-card-only hand (no made hand yet)', () => {
    expect(evaluateMississippiStudMadeHand([card('SPADE', 8), card('HEART', 3)])).toBeNull();
  });

  it('treats a low pair (2–5) as made but not pay-table eligible', () => {
    const result = evaluateMississippiStudMadeHand([card('SPADE', 4), card('HEART', 4)]);
    expect(result).toEqual({ rank: PokerHand.OnePair, paytableEligible: false });
  });

  it('treats a pair of 6s as a paying pair', () => {
    const result = evaluateMississippiStudMadeHand([card('SPADE', 6), card('HEART', 6)]);
    expect(result).toEqual({ rank: PokerHand.OnePair, paytableEligible: true });
  });

  it('treats a pair of aces (value 1) as a paying pair', () => {
    const result = evaluateMississippiStudMadeHand([card('SPADE', 1), card('HEART', 1)]);
    expect(result).toEqual({ rank: PokerHand.OnePair, paytableEligible: true });
  });

  it('names a made pair from a hole card plus a revealed community card', () => {
    const result = evaluateMississippiStudMadeHand([card('SPADE', 12), card('HEART', 3), card('CLOVER', 12)]);
    expect(result).toEqual({ rank: PokerHand.OnePair, paytableEligible: true });
  });

  it('names trips once a third matching card is revealed', () => {
    const result = evaluateMississippiStudMadeHand([card('SPADE', 12), card('HEART', 12), card('CLOVER', 12)]);
    expect(result).toEqual({ rank: PokerHand.ThreeOfAKind, paytableEligible: true });
  });

  it('detects a full five-card flush as pay-table eligible', () => {
    const result = evaluateMississippiStudMadeHand([
      card('SPADE', 2),
      card('SPADE', 5),
      card('SPADE', 8),
      card('SPADE', 11),
      card('SPADE', 13),
    ]);
    expect(result).toEqual({ rank: PokerHand.Flush, paytableEligible: true });
  });
});
