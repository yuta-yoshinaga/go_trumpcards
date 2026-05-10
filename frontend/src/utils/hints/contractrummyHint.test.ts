import { describe, expect, it } from 'vitest';
import type { ContractRummyResponse } from '../../types/card';
import { getContractRummyHint } from './contractrummyHint';

describe('getContractRummyHint', () => {
  it('returns null for null state', () => {
    expect(getContractRummyHint(null)).toBeNull();
  });

  it('returns null for any state (stub)', () => {
    const state: ContractRummyResponse = {
      players: [],
      phase: 0,
      roundNumber: 1,
      totalRounds: 7,
      currentPlayerIdx: 0,
      discardTop: null,
      drawPileCount: 60,
      gameEndFlag: false,
      winnerIdx: -1,
      roundWinnerIdx: -1,
      contractSlots: [],
      config: { cpuDifficulty: 1, failContractPenalty: 25 },
      message: '',
    };
    expect(getContractRummyHint(state)).toBeNull();
  });
});
