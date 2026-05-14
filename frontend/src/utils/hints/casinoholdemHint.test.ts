import { describe, expect, it } from 'vitest';
import type { CasinoHoldemResponse } from '../../types/card';
import { CasinoHoldemPhase } from '../../types/phases';
import { getCasinoHoldemHint } from './casinoholdemHint';

/** Build a CasinoHoldemResponse with sensible defaults so each test overrides
 * only the fields it cares about. */
function makeState(overrides: Partial<CasinoHoldemResponse> = {}): CasinoHoldemResponse {
  return {
    playerHand: [
      { design: 'SPADE', value: 7 },
      { design: 'HEART', value: 7 },
    ],
    dealerHand: [],
    community: [
      { design: 'DIAMOND', value: 3 },
      { design: 'CLOVER', value: 9 },
      { design: 'SPADE', value: 11 },
    ],
    phase: CasinoHoldemPhase.FLOP,
    chips: 1000,
    anteBet: 100,
    bonusBet: 0,
    callBet: 0,
    result: 0,
    dealerQualify: false,
    antePayout: 0,
    callPayout: 0,
    bonusPayout: 0,
    totalPayout: 0,
    playerHandRank: 1, // OnePair by default
    dealerHandRank: 0,
    message: '',
    ...overrides,
  };
}

describe('getCasinoHoldemHint', () => {
  it('returns call hint for pair-or-better at FLOP', () => {
    const hint = getCasinoHoldemHint(makeState({ playerHandRank: 1 }));
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toBe('call');
    expect(hint?.confidence).toBe('strong');
  });

  it('returns fold hint for High Card at FLOP', () => {
    const hint = getCasinoHoldemHint(makeState({ playerHandRank: 0 }));
    expect(hint).not.toBeNull();
    expect(hint?.targetAction).toBe('fold');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns null for BET phase', () => {
    const hint = getCasinoHoldemHint(makeState({ phase: CasinoHoldemPhase.BET }));
    expect(hint).toBeNull();
  });

  it('returns null for END phase', () => {
    const hint = getCasinoHoldemHint(makeState({ phase: CasinoHoldemPhase.END }));
    expect(hint).toBeNull();
  });

  it('returns null when no hole cards have been dealt', () => {
    const hint = getCasinoHoldemHint(makeState({ playerHand: [] }));
    expect(hint).toBeNull();
  });

  it('returns strong call for a Flush', () => {
    const hint = getCasinoHoldemHint(makeState({ playerHandRank: 5 }));
    expect(hint?.targetAction).toBe('call');
    expect(hint?.confidence).toBe('strong');
  });
});
