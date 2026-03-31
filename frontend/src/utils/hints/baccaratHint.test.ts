import { describe, expect, it } from 'vitest';
import type { BaccaratResponse } from '../../types/card';
import { getBaccaratHint } from './baccaratHint';

function makeState(overrides: Partial<BaccaratResponse> = {}): BaccaratResponse {
  return {
    playerHand: [],
    bankerHand: [],
    playerHandValue: 0,
    bankerHandValue: 0,
    phase: 1,
    chips: 1000,
    betAmount: 100,
    betType: 0,
    result: 0,
    payout: 0,
    history: [],
    playerPairBet: 0,
    bankerPairBet: 0,
    sideBetResults: [],
    message: '',
    ...overrides,
  };
}

describe('getBaccaratHint', () => {
  it('returns null when phase is not BET', () => {
    expect(getBaccaratHint(makeState({ phase: 2 }))).toBeNull();
  });

  it('returns banker hint during bet phase with no history', () => {
    const result = getBaccaratHint(makeState({ phase: 1, history: [] }));
    expect(result).toEqual({
      targetAction: 'banker',
      reason: 'hintReason.bankerBestOdds',
      confidence: 'strong',
    });
  });

  it('returns banker hint when history has mixed results', () => {
    const result = getBaccaratHint(makeState({ phase: 1, history: [0, 1, 2, 1, 0] }));
    expect(result).toEqual({
      targetAction: 'banker',
      reason: 'hintReason.bankerBestOdds',
      confidence: 'strong',
    });
  });

  it('returns avoid-tie hint when betType is TIE', () => {
    const result = getBaccaratHint(makeState({ phase: 1, betType: 2 }));
    expect(result).toEqual({
      targetAction: 'banker',
      reason: 'hintReason.avoidTie',
      confidence: 'strong',
    });
  });

  it('returns strong confidence banker hint when betType is PLAYER', () => {
    const result = getBaccaratHint(makeState({ phase: 1, betType: 0 }));
    expect(result).toEqual({
      targetAction: 'banker',
      reason: 'hintReason.bankerBestOdds',
      confidence: 'strong',
    });
  });

  it('returns null for banker bet (already optimal)', () => {
    const result = getBaccaratHint(makeState({ phase: 1, betType: 1 }));
    expect(result).toBeNull();
  });
});
