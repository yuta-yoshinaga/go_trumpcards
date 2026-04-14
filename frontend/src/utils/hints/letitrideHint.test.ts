import { describe, expect, it } from 'vitest';
import type { Card, LetItRideResponse } from '../../types/card';
import { LetItRidePhase } from '../../types/phases';
import { getLetitrideHint } from './letitrideHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<LetItRideResponse> = {}): LetItRideResponse {
  return {
    playerHand: [card('SPADE', 5), card('HEART', 8), card('DIAMOND', 3)],
    communityCards: [],
    phase: LetItRidePhase.FIRST_DECISION,
    chips: 1000,
    betAmount: 100,
    bet1Active: true,
    bet2Active: true,
    bet3Active: true,
    result: 0,
    handRank: 0,
    bet1Payout: 0,
    bet2Payout: 0,
    bet3Payout: 0,
    totalPayout: 0,
    message: '',
    ...overrides,
  };
}

describe('getLetitrideHint', () => {
  it('returns null for BET phase', () => {
    expect(getLetitrideHint(makeState({ phase: LetItRidePhase.BET }))).toBeNull();
  });

  it('returns null for END phase', () => {
    expect(getLetitrideHint(makeState({ phase: LetItRidePhase.END }))).toBeNull();
  });

  it('returns null when player hand is empty', () => {
    expect(getLetitrideHint(makeState({ playerHand: [] }))).toBeNull();
  });

  it('recommends letitride (strong) for three of a kind or better', () => {
    const hint = getLetitrideHint(makeState({ handRank: 3 }));
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.strongHand');
  });

  it('recommends letitride (strong) for high hand ranks (e.g. royal flush = 9)', () => {
    const hint = getLetitrideHint(makeState({ handRank: 9 }));
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('strong');
  });

  it('recommends letitride (moderate) for pair of tens (value=10)', () => {
    const hint = getLetitrideHint(
      makeState({
        handRank: 1,
        playerHand: [card('SPADE', 10), card('HEART', 10), card('DIAMOND', 3)],
      }),
    );
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.pairTensOrBetter');
  });

  it('recommends letitride (moderate) for pair of aces (value=1)', () => {
    const hint = getLetitrideHint(
      makeState({
        handRank: 1,
        playerHand: [card('SPADE', 1), card('HEART', 1), card('DIAMOND', 5)],
      }),
    );
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.pairTensOrBetter');
  });

  it('recommends letitride (moderate) for pair of jacks (value=11)', () => {
    const hint = getLetitrideHint(
      makeState({
        handRank: 1,
        playerHand: [card('SPADE', 11), card('HEART', 11), card('DIAMOND', 2)],
      }),
    );
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.pairTensOrBetter');
  });

  it('recommends pull for pair of nines (below 10, not ace)', () => {
    const hint = getLetitrideHint(
      makeState({
        handRank: 1,
        playerHand: [card('SPADE', 9), card('HEART', 9), card('DIAMOND', 5)],
      }),
    );
    expect(hint?.targetAction).toBe('pull');
  });

  it('recommends letitride (moderate) for three to a royal flush (same suit, all royal values)', () => {
    const hint = getLetitrideHint(
      makeState({
        handRank: 0,
        playerHand: [card('SPADE', 1), card('SPADE', 10), card('SPADE', 11)],
      }),
    );
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.threeToRoyalFlush');
  });

  it('recommends pull when royal values but different suits', () => {
    const hint = getLetitrideHint(
      makeState({
        handRank: 0,
        playerHand: [card('SPADE', 1), card('HEART', 10), card('DIAMOND', 11)],
      }),
    );
    expect(hint?.targetAction).toBe('pull');
  });

  it('recommends pull when same suit but non-royal values', () => {
    const hint = getLetitrideHint(
      makeState({
        handRank: 0,
        playerHand: [card('SPADE', 2), card('SPADE', 5), card('SPADE', 7)],
      }),
    );
    expect(hint?.targetAction).toBe('pull');
  });

  it('recommends pull for weak high-card hand', () => {
    const hint = getLetitrideHint(makeState({ handRank: 0 }));
    expect(hint?.targetAction).toBe('pull');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.weakHand');
  });

  it('also works in SECOND_DECISION phase', () => {
    const hint = getLetitrideHint(makeState({ phase: LetItRidePhase.SECOND_DECISION, handRank: 3 }));
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('strong');
  });

  it('returns pull when hand has fewer than 3 cards and handRank=0', () => {
    const hint = getLetitrideHint(
      makeState({
        handRank: 0,
        playerHand: [card('SPADE', 1), card('SPADE', 10)],
      }),
    );
    // cards.length < 3 → hasThreeToRoyalFlush returns false → pull
    expect(hint?.targetAction).toBe('pull');
  });
});
