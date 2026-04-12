import { describe, expect, it } from 'vitest';
import type { Card, CaribbeanStudResponse } from '../../types/card';
import { CaribbeanStudPhase } from '../../types/phases';
import { getCaribbeanStudHint } from './caribbeanstudHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<CaribbeanStudResponse> = {}): CaribbeanStudResponse {
  return {
    playerHand: [card('SPADE', 2), card('HEART', 5), card('DIAMOND', 8), card('CLOVER', 10), card('SPADE', 12)],
    dealerHand: [],
    phase: CaribbeanStudPhase.ACTION,
    chips: 1000,
    anteBet: 100,
    jackpotBet: 0,
    playBet: 0,
    result: 0,
    antePayout: 0,
    playPayout: 0,
    jackpotPayout: 0,
    totalPayout: 0,
    dealerQualified: false,
    playerHandRank: 0,
    dealerHandRank: 0,
    message: '',
    ...overrides,
  };
}

describe('getCaribbeanStudHint', () => {
  it('returns null outside action phase', () => {
    expect(getCaribbeanStudHint(makeState({ phase: CaribbeanStudPhase.BET }))).toBeNull();
    expect(getCaribbeanStudHint(makeState({ phase: CaribbeanStudPhase.END }))).toBeNull();
  });

  it('recommends play with pair or better (strong)', () => {
    const hint = getCaribbeanStudHint(makeState({ playerHandRank: 1 }));
    expect(hint?.targetAction).toBe('play');
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.pairOrBetter');
  });

  it('recommends play with Ace-King high (moderate)', () => {
    const state = makeState({
      playerHand: [card('SPADE', 1), card('HEART', 13), card('DIAMOND', 8), card('CLOVER', 5), card('SPADE', 3)],
      playerHandRank: 0,
    });
    const hint = getCaribbeanStudHint(state);
    expect(hint?.targetAction).toBe('play');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.aceKingHigh');
  });

  it('recommends fold with weak high-card hand', () => {
    const hint = getCaribbeanStudHint(makeState({ playerHandRank: 0 }));
    expect(hint?.targetAction).toBe('fold');
    expect(hint?.reason).toBe('hint.weakHand');
  });

  it('returns null when player hand is empty', () => {
    expect(getCaribbeanStudHint(makeState({ playerHand: [] }))).toBeNull();
  });
});
