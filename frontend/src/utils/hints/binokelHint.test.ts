import { describe, expect, it } from 'vitest';
import type { BinokelConfig, BinokelResponse } from '../../types/card';
import { BinokelPhase } from '../../types/phases';
import { getBinokelHint } from './binokelHint';

const defaultConfig: BinokelConfig = { cpuDifficulty: 0, pointLimit: 1500 };

function makeState(overrides: Partial<BinokelResponse> = {}): BinokelResponse {
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
    scores: [0, 0, 0],
    gameEndFlag: false,
    winnerPlayer: -1,
    leadPlayerIdx: 0,
    playerMelds: [[], [], []],
    dabb: [],
    dabbDiscarded: [],
    message: '',
    config: defaultConfig,
    ...overrides,
  };
}

describe('getBinokelHint', () => {
  it('returns null when game has ended', () => {
    expect(getBinokelHint(makeState({ gameEndFlag: true, hint: { reason: 'hint.playLow' } }))).toBeNull();
  });

  it('returns null when state has no hint', () => {
    expect(getBinokelHint(makeState())).toBeNull();
  });

  it('surfaces server hint_play reason and maps to hintReason namespace', () => {
    const hint = getBinokelHint(makeState({ hint: { cardIndex: 2, reason: 'hint_play' } }));
    expect(hint).toEqual({
      targetAction: 'play',
      reason: 'hintReason.hint_play',
      confidence: 'strong',
    });
  });

  it('maps pass hint to pass action', () => {
    const hint = getBinokelHint(makeState({ hint: { pass: true, reason: 'hint_pass' } }));
    expect(hint?.targetAction).toBe('pass');
    expect(hint?.reason).toBe('hintReason.hint_pass');
  });

  it('treats bidAmount 0 as pass action without outputting 0', () => {
    const hint = getBinokelHint(makeState({ hint: { bidAmount: 0, reason: 'hint_pass' } }));
    expect(hint?.targetAction).toBe('pass');
    expect(hint?.reason).toBe('hintReason.hint_pass');
  });

  it('maps bid hint to bid action', () => {
    const hint = getBinokelHint(makeState({ hint: { bidAmount: 160, reason: 'hint_bid' } }));
    expect(hint?.targetAction).toBe('bid');
    expect(hint?.reason).toBe('hintReason.hint_bid');
  });

  it('maps trump hint to trump action', () => {
    const hint = getBinokelHint(makeState({ hint: { suit: 1, reason: 'hint_trump' } }));
    expect(hint?.targetAction).toBe('trump');
    expect(hint?.reason).toBe('hintReason.hint_trump');
  });

  it('maps Dabb phase hint to discard action', () => {
    const hint = getBinokelHint(makeState({ phase: BinokelPhase.DABB, hint: { reason: 'hint_dabb' } }));
    expect(hint?.targetAction).toBe('discard');
    expect(hint?.reason).toBe('hintReason.hint_dabb');
  });

  it('falls back to the raw reason for unknown keys', () => {
    const hint = getBinokelHint(makeState({ hint: { reason: 'custom_reason' } }));
    expect(hint?.reason).toBe('custom_reason');
  });
});
