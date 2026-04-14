import { describe, expect, it } from 'vitest';
import type { Card, LetItRideResponse } from '../../types/card';
import { LetItRidePhase } from '../../types/phases';
import { getLetitrideHint } from './letitrideHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });
const masked = (): { design: ''; value: 0 } => ({ design: '', value: 0 });

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

  it('recommends letitride (strong) for three of a kind in player hand', () => {
    const hint = getLetitrideHint(
      makeState({ playerHand: [card('SPADE', 7), card('HEART', 7), card('DIAMOND', 7)] }),
    );
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.strongHand');
  });

  it('recommends letitride (strong) for three of a kind regardless of handRank value', () => {
    // handRank is always 0 during decision phases; the hint evaluates cards directly
    const hint = getLetitrideHint(
      makeState({ playerHand: [card('SPADE', 1), card('HEART', 1), card('DIAMOND', 1)] }),
    );
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

  it('also works in SECOND_DECISION phase (three of a kind via player cards)', () => {
    const hint = getLetitrideHint(
      makeState({
        phase: LetItRidePhase.SECOND_DECISION,
        playerHand: [card('SPADE', 7), card('HEART', 7), card('DIAMOND', 7)],
      }),
    );
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('strong');
  });

  it('returns pull when fewer than 3 royal-suit cards available', () => {
    const hint = getLetitrideHint(
      makeState({
        handRank: 0,
        playerHand: [card('SPADE', 1), card('SPADE', 10)],
      }),
    );
    // only 2 royal cards of the same suit → hasRoyalFlushDraw returns false → pull
    expect(hint?.targetAction).toBe('pull');
  });

  // SECOND_DECISION community card tests
  it('SECOND_DECISION: community card completes three of a kind', () => {
    const hint = getLetitrideHint(
      makeState({
        phase: LetItRidePhase.SECOND_DECISION,
        playerHand: [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 3)],
        communityCards: [card('CLUB', 5), masked()],
      }),
    );
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.strongHand');
  });

  it('SECOND_DECISION: community card creates pair of tens', () => {
    const hint = getLetitrideHint(
      makeState({
        phase: LetItRidePhase.SECOND_DECISION,
        playerHand: [card('SPADE', 10), card('HEART', 3), card('DIAMOND', 7)],
        communityCards: [card('CLUB', 10), masked()],
      }),
    );
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.pairTensOrBetter');
  });

  it('SECOND_DECISION: community card completes royal flush draw', () => {
    const hint = getLetitrideHint(
      makeState({
        phase: LetItRidePhase.SECOND_DECISION,
        // Player has only 2 spade royals — not enough alone
        playerHand: [card('SPADE', 1), card('SPADE', 13), card('DIAMOND', 4)],
        // Community reveals Q♠ → three spade royals (A, K, Q) total
        communityCards: [card('SPADE', 12), masked()],
      }),
    );
    expect(hint?.targetAction).toBe('letitride');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.threeToRoyalFlush');
  });

  it('SECOND_DECISION: masked community card is ignored', () => {
    const hint = getLetitrideHint(
      makeState({
        phase: LetItRidePhase.SECOND_DECISION,
        playerHand: [card('SPADE', 5), card('HEART', 8), card('DIAMOND', 3)],
        communityCards: [masked(), masked()],
      }),
    );
    expect(hint?.targetAction).toBe('pull');
  });

  it('SECOND_DECISION: weak hand with revealed community card → pull', () => {
    const hint = getLetitrideHint(
      makeState({
        phase: LetItRidePhase.SECOND_DECISION,
        playerHand: [card('SPADE', 2), card('HEART', 6), card('DIAMOND', 9)],
        communityCards: [card('CLUB', 4), masked()],
      }),
    );
    expect(hint?.targetAction).toBe('pull');
  });
});
