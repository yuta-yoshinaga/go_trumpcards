import { describe, expect, it } from 'vitest';
import type { UltimateTexasHoldemResponse } from '../../types/card';
import { UltimateTexasHoldemPhase } from '../../types/phases';
import { getUltimateTexasHoldemHint } from './ultimateTexasHoldemHint';

function baseState(overrides: Partial<UltimateTexasHoldemResponse> = {}): UltimateTexasHoldemResponse {
  return {
    playerHand: [],
    dealerHand: [],
    community: [],
    phase: UltimateTexasHoldemPhase.PRE_FLOP,
    chips: 1000,
    anteBet: 100,
    blindBet: 100,
    tripsBet: 0,
    playBet: 0,
    folded: false,
    result: 0,
    dealerQualified: false,
    antePayout: 0,
    blindPayout: 0,
    playPayout: 0,
    tripsPayout: 0,
    totalPayout: 0,
    playerHandRank: 0,
    dealerHandRank: 0,
    message: '',
    ...overrides,
  };
}

describe('getUltimateTexasHoldemHint', () => {
  it('returns null when hand has fewer than two cards', () => {
    expect(getUltimateTexasHoldemHint(baseState({ playerHand: [] }))).toBeNull();
  });

  it('recommends play with a pocket pair pre-flop', () => {
    const state = baseState({
      playerHand: [
        { design: 'SPADE', value: 13 },
        { design: 'HEART', value: 13 },
      ],
    });
    const hint = getUltimateTexasHoldemHint(state);
    expect(hint?.targetAction).toBe('play');
    expect(hint?.reason).toBe('hint.pocketPair');
    expect(hint?.confidence).toBe('strong');
  });

  it('strongly recommends play with a suited Ace pre-flop', () => {
    const state = baseState({
      playerHand: [
        { design: 'SPADE', value: 1 },
        { design: 'SPADE', value: 5 },
      ],
    });
    const hint = getUltimateTexasHoldemHint(state);
    expect(hint?.targetAction).toBe('play');
    expect(hint?.reason).toBe('hint.suitedBroadway');
  });

  it('recommends play with off-suit Ace pre-flop', () => {
    const state = baseState({
      playerHand: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 7 },
      ],
    });
    const hint = getUltimateTexasHoldemHint(state);
    expect(hint?.reason).toBe('hint.acePlay');
  });

  it('recommends play with off-suit Broadway pre-flop', () => {
    const state = baseState({
      playerHand: [
        { design: 'SPADE', value: 13 },
        { design: 'HEART', value: 11 },
      ],
    });
    const hint = getUltimateTexasHoldemHint(state);
    expect(hint?.targetAction).toBe('play');
    expect(hint?.reason).toBe('hint.broadwayCards');
  });

  it('recommends suited Broadway pre-flop', () => {
    const state = baseState({
      playerHand: [
        { design: 'SPADE', value: 13 },
        { design: 'SPADE', value: 11 },
      ],
    });
    const hint = getUltimateTexasHoldemHint(state);
    expect(hint?.reason).toBe('hint.suitedBroadway');
  });

  it('recommends check on weak holdings pre-flop', () => {
    const state = baseState({
      playerHand: [
        { design: 'SPADE', value: 3 },
        { design: 'HEART', value: 9 },
      ],
    });
    const hint = getUltimateTexasHoldemHint(state);
    expect(hint?.targetAction).toBe('check');
    expect(hint?.reason).toBe('hint.weakHand');
  });

  it('recommends raise (Play 2x) on flop with a pair', () => {
    const state = baseState({
      phase: UltimateTexasHoldemPhase.FLOP,
      playerHand: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 6 },
      ],
      playerHandRank: 1,
    });
    expect(getUltimateTexasHoldemHint(state)?.targetAction).toBe('raise');
  });

  it('recommends check on flop without a pair', () => {
    const state = baseState({
      phase: UltimateTexasHoldemPhase.FLOP,
      playerHand: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 8 },
      ],
      playerHandRank: 0,
    });
    expect(getUltimateTexasHoldemHint(state)?.targetAction).toBe('check');
  });

  it('recommends play (1x) on river with a made hand', () => {
    const state = baseState({
      phase: UltimateTexasHoldemPhase.RIVER,
      playerHand: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 5 },
      ],
      playerHandRank: 1,
    });
    const hint = getUltimateTexasHoldemHint(state);
    expect(hint?.targetAction).toBe('play');
    expect(hint?.reason).toBe('hint.madeHandRiver');
  });

  it('recommends fold on river when nothing connects', () => {
    const state = baseState({
      phase: UltimateTexasHoldemPhase.RIVER,
      playerHand: [
        { design: 'SPADE', value: 4 },
        { design: 'HEART', value: 9 },
      ],
      playerHandRank: 0,
    });
    const hint = getUltimateTexasHoldemHint(state);
    expect(hint?.targetAction).toBe('fold');
    expect(hint?.reason).toBe('hint.weakHand');
  });

  it('returns null outside decision phases', () => {
    const state = baseState({
      phase: UltimateTexasHoldemPhase.END,
      playerHand: [
        { design: 'SPADE', value: 5 },
        { design: 'HEART', value: 6 },
      ],
    });
    expect(getUltimateTexasHoldemHint(state)).toBeNull();
  });
});
