import { describe, expect, it } from 'vitest';
import type { Card, OasisPokerResponse } from '../../types/card';
import { OasisPokerPhase } from '../../types/phases';
import { getOasisPokerHint } from './oasispokerHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<OasisPokerResponse> = {}): OasisPokerResponse {
  return {
    playerHand: [card('SPADE', 2), card('HEART', 5), card('DIAMOND', 8), card('CLOVER', 10), card('SPADE', 12)],
    dealerHand: [],
    phase: OasisPokerPhase.ACTION,
    chips: 1000,
    anteBet: 100,
    jackpotBet: 0,
    exchangeCount: 0,
    exchangeFee: 0,
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

describe('getOasisPokerHint', () => {
  it('returns null in bet and end phases', () => {
    expect(getOasisPokerHint(makeState({ phase: OasisPokerPhase.BET }))).toBeNull();
    expect(getOasisPokerHint(makeState({ phase: OasisPokerPhase.END }))).toBeNull();
  });

  it('returns null when player hand is empty', () => {
    expect(getOasisPokerHint(makeState({ playerHand: [] }))).toBeNull();
  });

  // --- Action phase ---
  it('recommends play with pair or better (strong)', () => {
    const hint = getOasisPokerHint(makeState({ phase: OasisPokerPhase.ACTION, playerHandRank: 1 }));
    expect(hint?.targetAction).toBe('play');
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.pairOrBetter');
  });

  it('recommends play with Ace-King high (moderate)', () => {
    const state = makeState({
      phase: OasisPokerPhase.ACTION,
      playerHand: [card('SPADE', 1), card('HEART', 13), card('DIAMOND', 8), card('CLOVER', 5), card('SPADE', 3)],
      playerHandRank: 0,
    });
    const hint = getOasisPokerHint(state);
    expect(hint?.targetAction).toBe('play');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.aceKingHigh');
  });

  it('recommends fold with a weak high-card hand', () => {
    const hint = getOasisPokerHint(makeState({ phase: OasisPokerPhase.ACTION, playerHandRank: 0 }));
    expect(hint?.targetAction).toBe('fold');
    expect(hint?.reason).toBe('hint.weakHand');
  });

  // --- Exchange phase ---
  it('recommends standing with a made hand of pair or better (strong)', () => {
    const hint = getOasisPokerHint(makeState({ phase: OasisPokerPhase.EXCHANGE, playerHandRank: 1 }));
    expect(hint?.targetAction).toBe('stand');
    expect(hint?.confidence).toBe('strong');
    expect(hint?.reason).toBe('hint.exchangeKeep');
  });

  it('recommends exchanging a weak hand (moderate)', () => {
    const hint = getOasisPokerHint(makeState({ phase: OasisPokerPhase.EXCHANGE, playerHandRank: 0 }));
    expect(hint?.targetAction).toBe('exchange');
    expect(hint?.confidence).toBe('moderate');
    expect(hint?.reason).toBe('hint.exchangeImprove');
  });
});
