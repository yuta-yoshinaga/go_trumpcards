import { describe, expect, it } from 'vitest';
import type { Card, MississippiStudResponse } from '../../types/card';
import { MississippiStudPhase } from '../../types/phases';
import { getMississippiStudHint } from './mississippiStudHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<MississippiStudResponse> = {}): MississippiStudResponse {
  return {
    playerHand: [card('SPADE', 5), card('HEART', 8)],
    communityCards: [],
    communityRevealed: [false, false, false],
    phase: MississippiStudPhase.THIRD_STREET,
    chips: 1000,
    anteAmount: 100,
    streetMultipliers: [0, 0, 0],
    folded: false,
    totalBet: 100,
    result: 0,
    handRank: 0,
    payoutMultiplier: 0,
    antePayout: 0,
    streetPayouts: [0, 0, 0],
    totalPayout: 0,
    message: '',
    ...overrides,
  };
}

describe('getMississippiStudHint', () => {
  it('returns null for ANTE phase', () => {
    expect(getMississippiStudHint(makeState({ phase: MississippiStudPhase.ANTE }))).toBeNull();
  });

  it('returns null for END phase', () => {
    expect(getMississippiStudHint(makeState({ phase: MississippiStudPhase.END }))).toBeNull();
  });

  it('returns null when player hand is empty', () => {
    expect(getMississippiStudHint(makeState({ playerHand: [] }))).toBeNull();
  });

  it('recommends 3x raise on a qualifying pair (Jacks)', () => {
    const state = makeState({ playerHand: [card('SPADE', 11), card('HEART', 11)] });
    const hint = getMississippiStudHint(state);
    expect(hint).toEqual({ targetAction: 'play3x', reason: 'hint.raiseBig', confidence: 'strong' });
  });

  it('recommends 3x raise on a qualifying mid pair (8s)', () => {
    const state = makeState({ playerHand: [card('SPADE', 8), card('HEART', 8)] });
    expect(getMississippiStudHint(state)?.targetAction).toBe('play3x');
  });

  it('recommends 3x raise on Aces (qualifying)', () => {
    const state = makeState({ playerHand: [card('SPADE', 1), card('HEART', 1)] });
    expect(getMississippiStudHint(state)?.targetAction).toBe('play3x');
  });

  it('recommends 1x for two high cards at 3rd Street (Q + K)', () => {
    const state = makeState({ playerHand: [card('SPADE', 12), card('HEART', 13)] });
    expect(getMississippiStudHint(state)?.targetAction).toBe('play1x');
  });

  it('recommends 1x for a flush draw (4 to a flush, 4th Street)', () => {
    const state = makeState({
      phase: MississippiStudPhase.FOURTH_STREET,
      playerHand: [card('SPADE', 5), card('SPADE', 9)],
      communityCards: [card('SPADE', 2), card('SPADE', 4)],
      communityRevealed: [true, true, false],
    });
    expect(getMississippiStudHint(state)?.targetAction).toBe('play1x');
  });

  it('recommends 1x for an open-ended straight draw at 4th Street', () => {
    const state = makeState({
      phase: MississippiStudPhase.FOURTH_STREET,
      playerHand: [card('SPADE', 6), card('HEART', 7)],
      communityCards: [card('DIAMOND', 8), card('CLOVER', 9)],
      communityRevealed: [true, true, false],
    });
    expect(getMississippiStudHint(state)?.targetAction).toBe('play1x');
  });

  it('recommends fold for a weak hand (5-7 unsuited, 3rd Street)', () => {
    const state = makeState({ playerHand: [card('SPADE', 5), card('HEART', 7)] });
    expect(getMississippiStudHint(state)?.targetAction).toBe('fold');
  });

  it('recommends fold at 5th Street if no made hand and no draw', () => {
    const state = makeState({
      phase: MississippiStudPhase.FIFTH_STREET,
      playerHand: [card('SPADE', 5), card('HEART', 7)],
      communityCards: [card('DIAMOND', 2), card('CLOVER', 3)],
      communityRevealed: [true, true, false],
    });
    expect(getMississippiStudHint(state)?.targetAction).toBe('fold');
  });
});
