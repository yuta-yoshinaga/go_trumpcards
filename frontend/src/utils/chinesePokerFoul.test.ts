import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { chinesePokerIsFoul, cpEvalFiveCardHand, cpEvalThreeCardHand } from './chinesePokerFoul';

const S = 'SPADE';
const H = 'HEART';
const D = 'DIAMOND';
const C = 'CLOVER';

/** Build a card. */
function c(design: CardDesign, value: number): Card {
  return { design, value };
}

describe('cpEvalFiveCardHand', () => {
  it('detects a flush', () => {
    expect(cpEvalFiveCardHand([c(S, 2), c(S, 5), c(S, 8), c(S, 11), c(S, 13)])).toBe(5);
  });
  it('detects a full house', () => {
    expect(cpEvalFiveCardHand([c(S, 7), c(H, 7), c(D, 7), c(C, 4), c(S, 4)])).toBe(6);
  });
  it('detects a wheel straight (A-2-3-4-5)', () => {
    expect(cpEvalFiveCardHand([c(S, 1), c(H, 2), c(D, 3), c(C, 4), c(S, 5)])).toBe(4);
  });
  it('detects a broadway straight (A-10-J-Q-K)', () => {
    expect(cpEvalFiveCardHand([c(S, 1), c(H, 10), c(D, 11), c(C, 12), c(S, 13)])).toBe(4);
  });
  it('detects a straight flush', () => {
    expect(cpEvalFiveCardHand([c(S, 2), c(S, 3), c(S, 4), c(S, 5), c(S, 6)])).toBe(8);
  });
  it('detects a royal flush', () => {
    expect(cpEvalFiveCardHand([c(S, 1), c(S, 10), c(S, 11), c(S, 12), c(S, 13)])).toBe(9);
  });
  it('returns high card for a non-5-card hand', () => {
    expect(cpEvalFiveCardHand([c(S, 2), c(H, 3)])).toBe(0);
  });
});

describe('cpEvalThreeCardHand', () => {
  it('ranks trips above straight', () => {
    expect(cpEvalThreeCardHand([c(S, 9), c(H, 9), c(D, 9)])).toBe(5);
  });
  it('detects a 3-card straight', () => {
    expect(cpEvalThreeCardHand([c(S, 5), c(H, 6), c(D, 7)])).toBe(4);
  });
  it('detects a pair', () => {
    expect(cpEvalThreeCardHand([c(S, 5), c(H, 5), c(D, 9)])).toBe(2);
  });
  it('returns high card for a non-3-card hand', () => {
    expect(cpEvalThreeCardHand([c(S, 2)])).toBe(1);
  });
});

describe('chinesePokerIsFoul', () => {
  const front = [c(S, 2), c(H, 3), c(D, 5)]; // high card (5 high)

  it('returns false for a legal arrangement (back >= middle >= front)', () => {
    const middle = [c(S, 6), c(H, 6), c(D, 9), c(C, 10), c(S, 12)]; // one pair
    const back = [c(S, 8), c(H, 8), c(D, 8), c(C, 2), c(S, 4)]; // trips
    expect(chinesePokerIsFoul(front, middle, back)).toBe(false);
  });

  it('returns true when back is weaker than middle (category)', () => {
    const middle = [c(S, 8), c(H, 8), c(D, 8), c(C, 2), c(S, 4)]; // trips
    const back = [c(S, 6), c(H, 6), c(D, 9), c(C, 10), c(S, 12)]; // one pair
    expect(chinesePokerIsFoul(front, middle, back)).toBe(true);
  });

  it('returns true when back ties middle on category but loses the high-card tiebreak', () => {
    const middle = [c(S, 13), c(H, 13), c(D, 9), c(C, 10), c(S, 12)]; // pair of Kings
    const back = [c(S, 6), c(H, 6), c(D, 9), c(C, 10), c(S, 12)]; // pair of 6s
    expect(chinesePokerIsFoul(front, middle, back)).toBe(true);
  });

  it('returns true when front is stronger than middle', () => {
    const strongFront = [c(S, 9), c(H, 9), c(D, 9)]; // trips
    const middle = [c(S, 6), c(H, 6), c(D, 2), c(C, 10), c(S, 12)]; // one pair
    const back = [c(S, 8), c(H, 8), c(D, 8), c(C, 8), c(S, 4)]; // quads
    expect(chinesePokerIsFoul(strongFront, middle, back)).toBe(true);
  });

  it('returns true when front ties middle on rank but wins the kicker comparison', () => {
    // Front pair of Kings vs middle pair of Queens: same mapped rank (one pair),
    // front's higher kicker makes it stronger than middle → foul.
    const frontPairK = [c(S, 13), c(H, 13), c(S, 2)];
    const middle = [c(S, 12), c(H, 12), c(C, 5), c(D, 4), c(C, 3)]; // pair of Queens
    const back = [c(S, 8), c(H, 8), c(C, 8), c(D, 7), c(S, 6)]; // trips
    expect(chinesePokerIsFoul(frontPairK, middle, back)).toBe(true);
  });

  it('returns false when front ties middle on rank but loses the kicker comparison', () => {
    // Front pair of 3s vs middle pair of Queens: same mapped rank, front's lower
    // kicker keeps it weaker than middle → legal.
    const frontPair3 = [c(S, 3), c(H, 3), c(D, 2)];
    const middle = [c(S, 12), c(H, 12), c(C, 5), c(D, 4), c(C, 3)]; // pair of Queens
    const back = [c(S, 8), c(H, 8), c(C, 8), c(D, 7), c(S, 6)]; // trips
    expect(chinesePokerIsFoul(frontPair3, middle, back)).toBe(false);
  });

  it('returns false for incomplete (wrong-length) rows', () => {
    expect(chinesePokerIsFoul([c(S, 2)], [], [])).toBe(false);
  });
});
