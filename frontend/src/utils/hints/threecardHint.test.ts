import { describe, expect, it } from 'vitest';
import type { ThreeCardResponse } from '../../types/card';
import { getThreeCardHint } from './threecardHint';

function makeState(overrides: Partial<ThreeCardResponse> = {}): ThreeCardResponse {
  return {
    playerHand: [],
    dealerHand: [],
    phase: 2,
    chips: 1000,
    anteBet: 100,
    pairPlusBet: 0,
    playBet: 0,
    result: 0,
    antePayout: 0,
    playPayout: 0,
    anteBonusPayout: 0,
    pairPlusPayout: 0,
    totalPayout: 0,
    dealerQualified: false,
    playerHandRank: 1,
    dealerHandRank: 0,
    message: '',
    ...overrides,
  };
}

describe('getThreeCardHint', () => {
  it('returns null when phase is not ACTION', () => {
    expect(getThreeCardHint(makeState({ phase: 1 }))).toBeNull();
    expect(getThreeCardHint(makeState({ phase: 3 }))).toBeNull();
  });

  it('returns play hint for pair or better (rank >= 2)', () => {
    const result = getThreeCardHint(
      makeState({
        phase: 2,
        playerHandRank: 2,
        playerHand: [
          { design: 'HEART', value: 10 },
          { design: 'DIAMOND', value: 10 },
          { design: 'SPADE', value: 5 },
        ],
      }),
    );
    expect(result).toEqual({
      targetAction: 'play',
      reason: 'hintReason.strongHand',
      confidence: 'strong',
    });
  });

  it('returns play hint for three of a kind (rank 5)', () => {
    const result = getThreeCardHint(
      makeState({
        phase: 2,
        playerHandRank: 5,
        playerHand: [
          { design: 'HEART', value: 7 },
          { design: 'DIAMOND', value: 7 },
          { design: 'SPADE', value: 7 },
        ],
      }),
    );
    expect(result).toEqual({
      targetAction: 'play',
      reason: 'hintReason.strongHand',
      confidence: 'strong',
    });
  });

  it('returns play hint for straight flush (rank 6)', () => {
    const result = getThreeCardHint(
      makeState({
        phase: 2,
        playerHandRank: 6,
        playerHand: [
          { design: 'HEART', value: 5 },
          { design: 'HEART', value: 6 },
          { design: 'HEART', value: 7 },
        ],
      }),
    );
    expect(result).toEqual({
      targetAction: 'play',
      reason: 'hintReason.strongHand',
      confidence: 'strong',
    });
  });

  it('returns play hint for high card hand Q-7-4 (above Q-6-4)', () => {
    const result = getThreeCardHint(
      makeState({
        phase: 2,
        playerHandRank: 1,
        playerHand: [
          { design: 'HEART', value: 12 },
          { design: 'DIAMOND', value: 7 },
          { design: 'SPADE', value: 4 },
        ],
      }),
    );
    expect(result).toEqual({
      targetAction: 'play',
      reason: 'hintReason.queenHighPlay',
      confidence: 'moderate',
    });
  });

  it('returns play hint for exactly Q-6-4 (boundary)', () => {
    const result = getThreeCardHint(
      makeState({
        phase: 2,
        playerHandRank: 1,
        playerHand: [
          { design: 'HEART', value: 12 },
          { design: 'DIAMOND', value: 6 },
          { design: 'SPADE', value: 4 },
        ],
      }),
    );
    expect(result).toEqual({
      targetAction: 'play',
      reason: 'hintReason.queenHighPlay',
      confidence: 'moderate',
    });
  });

  it('returns fold hint for Q-6-3 (below Q-6-4)', () => {
    const result = getThreeCardHint(
      makeState({
        phase: 2,
        playerHandRank: 1,
        playerHand: [
          { design: 'HEART', value: 12 },
          { design: 'DIAMOND', value: 6 },
          { design: 'SPADE', value: 3 },
        ],
      }),
    );
    expect(result).toEqual({
      targetAction: 'fold',
      reason: 'hintReason.weakHand',
      confidence: 'moderate',
    });
  });

  it('returns fold hint for J-high hand', () => {
    const result = getThreeCardHint(
      makeState({
        phase: 2,
        playerHandRank: 1,
        playerHand: [
          { design: 'HEART', value: 11 },
          { design: 'DIAMOND', value: 9 },
          { design: 'SPADE', value: 5 },
        ],
      }),
    );
    expect(result).toEqual({
      targetAction: 'fold',
      reason: 'hintReason.weakHand',
      confidence: 'moderate',
    });
  });

  it('returns play hint for K-high hand', () => {
    const result = getThreeCardHint(
      makeState({
        phase: 2,
        playerHandRank: 1,
        playerHand: [
          { design: 'HEART', value: 13 },
          { design: 'DIAMOND', value: 3 },
          { design: 'SPADE', value: 2 },
        ],
      }),
    );
    expect(result).toEqual({
      targetAction: 'play',
      reason: 'hintReason.queenHighPlay',
      confidence: 'moderate',
    });
  });

  it('returns play hint for A-high hand', () => {
    const result = getThreeCardHint(
      makeState({
        phase: 2,
        playerHandRank: 1,
        playerHand: [
          { design: 'HEART', value: 1 },
          { design: 'DIAMOND', value: 3 },
          { design: 'SPADE', value: 2 },
        ],
      }),
    );
    expect(result).toEqual({
      targetAction: 'play',
      reason: 'hintReason.queenHighPlay',
      confidence: 'moderate',
    });
  });

  it('returns fold hint when hand is empty', () => {
    const result = getThreeCardHint(
      makeState({
        phase: 2,
        playerHandRank: 1,
        playerHand: [],
      }),
    );
    expect(result).toEqual({
      targetAction: 'fold',
      reason: 'hintReason.weakHand',
      confidence: 'moderate',
    });
  });
});
