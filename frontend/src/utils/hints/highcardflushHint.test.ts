import { describe, expect, it } from 'vitest';
import type { HighCardFlushResponse } from '../../types/card';
import { getHighCardFlushHint } from './highcardflushHint';

function makeState(overrides: Partial<HighCardFlushResponse> = {}): HighCardFlushResponse {
  return {
    playerHand: [],
    dealerHand: [],
    phase: 2,
    chips: 1000,
    anteBet: 100,
    flushBonusBet: 0,
    straightFlushBet: 0,
    raiseBet: 0,
    result: 0,
    antePayout: 0,
    raisePayout: 0,
    flushBonusPayout: 0,
    straightFlushPayout: 0,
    totalPayout: 0,
    dealerQualified: false,
    playerFlushLen: 0,
    dealerFlushLen: 0,
    playerStraightFlushLen: 0,
    maxRaiseMultiplier: 1,
    message: '',
    ...overrides,
  };
}

describe('getHighCardFlushHint', () => {
  it('returns null when phase is not ACTION', () => {
    expect(getHighCardFlushHint(makeState({ phase: 1 }))).toBeNull();
    expect(getHighCardFlushHint(makeState({ phase: 3 }))).toBeNull();
  });

  it('returns raise3x for 6-card flush', () => {
    const r = getHighCardFlushHint(makeState({ phase: 2, playerFlushLen: 6 }));
    expect(r).toEqual({ targetAction: 'raise3x', reason: 'hintReason.raise3x', confidence: 'strong' });
  });

  it('returns raise3x for 7-card flush', () => {
    const r = getHighCardFlushHint(makeState({ phase: 2, playerFlushLen: 7 }));
    expect(r?.targetAction).toBe('raise3x');
  });

  it('returns raise2x for 5-card flush', () => {
    const r = getHighCardFlushHint(makeState({ phase: 2, playerFlushLen: 5 }));
    expect(r).toEqual({ targetAction: 'raise2x', reason: 'hintReason.raise2x', confidence: 'strong' });
  });

  it('returns raise1x for 4-card flush', () => {
    const r = getHighCardFlushHint(makeState({ phase: 2, playerFlushLen: 4 }));
    expect(r).toEqual({ targetAction: 'raise1x', reason: 'hintReason.raise1xHighFlush', confidence: 'strong' });
  });

  it('returns raise1x for 3-card flush with Queen-high', () => {
    const r = getHighCardFlushHint(
      makeState({
        phase: 2,
        playerFlushLen: 3,
        playerHand: [
          { design: 'SPADE', value: 5 },
          { design: 'SPADE', value: 8 },
          { design: 'SPADE', value: 12 },
          { design: 'HEART', value: 9 },
          { design: 'CLOVER', value: 4 },
          { design: 'CLOVER', value: 6 },
          { design: 'DIAMOND', value: 7 },
        ],
      }),
    );
    expect(r).toEqual({
      targetAction: 'raise1x',
      reason: 'hintReason.raise1xMarginal',
      confidence: 'moderate',
    });
  });

  it('returns raise1x for 3-card flush with Ace-high (Ace counts as 14)', () => {
    const r = getHighCardFlushHint(
      makeState({
        phase: 2,
        playerFlushLen: 3,
        playerHand: [
          { design: 'SPADE', value: 5 },
          { design: 'SPADE', value: 8 },
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 9 },
          { design: 'CLOVER', value: 4 },
          { design: 'CLOVER', value: 6 },
          { design: 'DIAMOND', value: 7 },
        ],
      }),
    );
    expect(r?.targetAction).toBe('raise1x');
  });

  it('returns fold for 3-card flush below Queen-high', () => {
    const r = getHighCardFlushHint(
      makeState({
        phase: 2,
        playerFlushLen: 3,
        playerHand: [
          { design: 'SPADE', value: 5 },
          { design: 'SPADE', value: 8 },
          { design: 'SPADE', value: 10 },
          { design: 'HEART', value: 9 },
          { design: 'CLOVER', value: 4 },
          { design: 'CLOVER', value: 6 },
          { design: 'DIAMOND', value: 7 },
        ],
      }),
    );
    expect(r).toEqual({ targetAction: 'fold', reason: 'hintReason.fold', confidence: 'moderate' });
  });

  it('returns fold for 2-card flush only', () => {
    const r = getHighCardFlushHint(makeState({ phase: 2, playerFlushLen: 2 }));
    expect(r?.targetAction).toBe('fold');
  });
});
