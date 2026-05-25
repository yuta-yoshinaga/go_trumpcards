import { describe, expect, it } from 'vitest';
import type { RussianPokerResponse } from '../../types/card';
import { RussianPokerPhase } from '../../types/phases';
import { getRussianPokerHint } from './russianpokerHint';

function makeState(overrides: Partial<RussianPokerResponse>): RussianPokerResponse {
  return {
    playerHand: [],
    dealerHand: [],
    phase: RussianPokerPhase.ACTION,
    chips: 1000,
    anteBet: 100,
    exchangeCount: 0,
    exchangeFee: 0,
    bought6th: false,
    buy6thFee: 0,
    forceExchanged: false,
    forceExchangeFee: 0,
    playBet: 0,
    result: 0,
    antePayout: 0,
    playPayout: 0,
    totalPayout: 0,
    dealerQualified: false,
    playerHandRank: 0,
    dealerHandRank: 0,
    message: '',
    ...overrides,
  };
}

describe('getRussianPokerHint', () => {
  it('returns null for non-action phases', () => {
    expect(getRussianPokerHint(makeState({ phase: RussianPokerPhase.BET }))).toBeNull();
    expect(getRussianPokerHint(makeState({ phase: RussianPokerPhase.SELECT }))).toBeNull();
    expect(getRussianPokerHint(makeState({ phase: RussianPokerPhase.END }))).toBeNull();
    expect(getRussianPokerHint(makeState({ phase: RussianPokerPhase.FORCE_QUALIFY }))).toBeNull();
  });

  it('returns null when playerHand is empty', () => {
    expect(getRussianPokerHint(makeState({ playerHand: [] }))).toBeNull();
  });

  it('recommends play with pair or better', () => {
    const state = makeState({
      phase: RussianPokerPhase.ACTION,
      playerHand: [
        { design: 'spade', value: 2 },
        { design: 'clover', value: 2 },
        { design: 'heart', value: 5 },
        { design: 'diamond', value: 7 },
        { design: 'spade', value: 9 },
      ],
      playerHandRank: 1,
    });
    const hint = getRussianPokerHint(state);
    expect(hint).not.toBeNull();
    expect(hint!.targetAction).toBe('play');
    expect(hint!.confidence).toBe('strong');
  });

  it('recommends play with Ace-King high', () => {
    const state = makeState({
      phase: RussianPokerPhase.ACTION,
      playerHand: [
        { design: 'spade', value: 1 },
        { design: 'clover', value: 13 },
        { design: 'heart', value: 5 },
        { design: 'diamond', value: 7 },
        { design: 'spade', value: 9 },
      ],
      playerHandRank: 0,
    });
    const hint = getRussianPokerHint(state);
    expect(hint).not.toBeNull();
    expect(hint!.targetAction).toBe('play');
    expect(hint!.confidence).toBe('moderate');
  });

  it('recommends fold with weak hand', () => {
    const state = makeState({
      phase: RussianPokerPhase.ACTION,
      playerHand: [
        { design: 'spade', value: 2 },
        { design: 'clover', value: 5 },
        { design: 'heart', value: 7 },
        { design: 'diamond', value: 9 },
        { design: 'spade', value: 11 },
      ],
      playerHandRank: 0,
    });
    const hint = getRussianPokerHint(state);
    expect(hint).not.toBeNull();
    expect(hint!.targetAction).toBe('fold');
  });

  it('works in PostAction phase', () => {
    const state = makeState({
      phase: RussianPokerPhase.POST_ACTION,
      playerHand: [
        { design: 'spade', value: 2 },
        { design: 'clover', value: 2 },
        { design: 'heart', value: 5 },
        { design: 'diamond', value: 7 },
        { design: 'spade', value: 9 },
      ],
      playerHandRank: 1,
    });
    const hint = getRussianPokerHint(state);
    expect(hint).not.toBeNull();
    expect(hint!.targetAction).toBe('play');
  });
});
