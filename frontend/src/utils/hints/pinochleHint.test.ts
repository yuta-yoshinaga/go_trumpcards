import { describe, expect, it } from 'vitest';
import type { PinochleConfig, PinochleResponse } from '../../types/card';
import { getPinochleHint } from './pinochleHint';

const defaultConfig: PinochleConfig = { cpuDifficulty: 0, pointLimit: 1500 };

function makeState(overrides: Partial<PinochleResponse> = {}): PinochleResponse {
  return {
    players: [],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 0,
    trumpSuit: 0,
    highestBid: 0,
    highestBidder: -1,
    currentTrick: [],
    teamScores: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    playerMelds: [[], [], [], []],
    message: '',
    config: defaultConfig,
    ...overrides,
  };
}

describe('getPinochleHint', () => {
  it('returns null when game has ended', () => {
    expect(getPinochleHint(makeState({ gameEndFlag: true, hint: { reason: 'hint.playLow' } }))).toBeNull();
  });

  it('returns null when state has no hint', () => {
    expect(getPinochleHint(makeState())).toBeNull();
  });

  it('surfaces server hint_play reason and maps to hintReason namespace', () => {
    const hint = getPinochleHint(makeState({ hint: { cardIndex: 2, reason: 'hint_play' } }));
    expect(hint).toEqual({
      targetAction: 'play',
      reason: 'hintReason.hint_play',
      confidence: 'strong',
    });
  });

  it('maps pass hint to pass action', () => {
    const hint = getPinochleHint(makeState({ hint: { pass: true, reason: 'hint_pass' } }));
    expect(hint?.targetAction).toBe('pass');
    expect(hint?.reason).toBe('hintReason.hint_pass');
  });

  it('maps bid hint to bid action', () => {
    const hint = getPinochleHint(makeState({ hint: { bidAmount: 30, reason: 'hint_bid' } }));
    expect(hint?.targetAction).toBe('bid');
    expect(hint?.reason).toBe('hintReason.hint_bid');
  });

  it('falls back to the raw reason for unknown keys', () => {
    const hint = getPinochleHint(makeState({ hint: { reason: 'custom_reason' } }));
    expect(hint?.reason).toBe('custom_reason');
  });
});
