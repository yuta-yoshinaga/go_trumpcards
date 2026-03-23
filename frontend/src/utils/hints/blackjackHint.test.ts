import { describe, expect, it } from 'vitest';
import {
  BJ_SUGGEST_DECLINE_INSURANCE,
  BJ_SUGGEST_DOUBLE,
  BJ_SUGGEST_DOUBLE_STAND,
  BJ_SUGGEST_HIT,
  BJ_SUGGEST_NONE,
  BJ_SUGGEST_SPLIT,
  BJ_SUGGEST_STAND,
  BJ_SUGGEST_SURRENDER,
} from '../../components/blackjack/bjConstants';
import type { BlackJackResponse } from '../../types/card';
import { getBlackjackHint } from './blackjackHint';

function makeState(overrides: Partial<BlackJackResponse> = {}): BlackJackResponse {
  return {
    dealer: { score: 0, cards: [], chips: 1000 },
    player: { score: 0, cards: [], chips: 1000 },
    hands: [],
    currentHandIdx: 0,
    phase: 4,
    insuranceBet: 0,
    insuranceAvailable: false,
    message: '',
    hintEnabled: true,
    suggestedAction: BJ_SUGGEST_NONE,
    deckCount: 6,
    dealerHitsSoft17: false,
    countingEnabled: false,
    cpuPlayerCount: 0,
    runningCount: 0,
    trueCount: 0,
    perfectPairsBet: 0,
    twentyOnePlus3Bet: 0,
    doubleAfterSplit: false,
    countingSystem: 0,
    deckPenetration: 75,
    multiHandCount: 1,
    surrenderRule: 0,
    ...overrides,
  };
}

describe('getBlackjackHint', () => {
  it('returns null when hintEnabled is false', () => {
    expect(getBlackjackHint(makeState({ hintEnabled: false, suggestedAction: BJ_SUGGEST_HIT }))).toBeNull();
  });

  it('returns null when suggestedAction is NONE', () => {
    expect(getBlackjackHint(makeState({ suggestedAction: BJ_SUGGEST_NONE }))).toBeNull();
  });

  it('maps HIT action', () => {
    const result = getBlackjackHint(makeState({ suggestedAction: BJ_SUGGEST_HIT }));
    expect(result).toEqual({ targetAction: 'hit', reason: 'hintReason.hitReason', confidence: 'strong' });
  });

  it('maps STAND action', () => {
    const result = getBlackjackHint(makeState({ suggestedAction: BJ_SUGGEST_STAND }));
    expect(result?.targetAction).toBe('stand');
  });

  it('maps DOUBLE action', () => {
    const result = getBlackjackHint(makeState({ suggestedAction: BJ_SUGGEST_DOUBLE }));
    expect(result?.targetAction).toBe('double');
  });

  it('maps DOUBLE_STAND action', () => {
    const result = getBlackjackHint(makeState({ suggestedAction: BJ_SUGGEST_DOUBLE_STAND }));
    expect(result?.targetAction).toBe('double');
  });

  it('maps SPLIT action', () => {
    const result = getBlackjackHint(makeState({ suggestedAction: BJ_SUGGEST_SPLIT }));
    expect(result?.targetAction).toBe('split');
  });

  it('maps SURRENDER action', () => {
    const result = getBlackjackHint(makeState({ suggestedAction: BJ_SUGGEST_SURRENDER }));
    expect(result?.targetAction).toBe('surrender');
  });

  it('maps DECLINE_INSURANCE action', () => {
    const result = getBlackjackHint(makeState({ suggestedAction: BJ_SUGGEST_DECLINE_INSURANCE }));
    expect(result?.targetAction).toBe('decline');
  });

  it('returns null for unknown suggestedAction', () => {
    expect(getBlackjackHint(makeState({ suggestedAction: 999 }))).toBeNull();
  });

  it('all mapped results have strong confidence', () => {
    for (const action of [BJ_SUGGEST_HIT, BJ_SUGGEST_STAND, BJ_SUGGEST_DOUBLE, BJ_SUGGEST_SPLIT]) {
      const result = getBlackjackHint(makeState({ suggestedAction: action }));
      expect(result?.confidence).toBe('strong');
    }
  });
});
