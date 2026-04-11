import { describe, expect, it } from 'vitest';
import type { TwoTenJackConfig, TwoTenJackResponse } from '../../types/card';
import { getTwoTenJackHint } from './twotenjackHint';

const defaultConfig: TwoTenJackConfig = { cpuDifficulty: 0, pointLimit: 1500 };

function makeState(overrides: Partial<TwoTenJackResponse> = {}): TwoTenJackResponse {
  return {
    players: [],
    phase: 1,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    declarerIdx: 0,
    trumpSuit: 1,
    currentTrick: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: defaultConfig,
    ...overrides,
  };
}

describe('getTwoTenJackHint', () => {
  it('returns null when game has ended', () => {
    expect(getTwoTenJackHint(makeState({ gameEndFlag: true, hint: { reason: 'hint.playStrong' } }))).toBeNull();
  });

  it('returns null when state has no hint', () => {
    expect(getTwoTenJackHint(makeState())).toBeNull();
  });

  it('surfaces server play hint and maps to hint namespace', () => {
    const hint = getTwoTenJackHint(makeState({ hint: { cardIndex: 3, reason: 'follow_suit' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.follow_suit', confidence: 'strong' });
  });

  it('maps declare-only hint to declare action', () => {
    const hint = getTwoTenJackHint(makeState({ hint: { trumpSuit: 2, reason: 'strategic_trump' } }));
    expect(hint?.targetAction).toBe('declare');
    expect(hint?.reason).toBe('hint.strategic_trump');
  });

  it('falls back to raw reason for unknown keys', () => {
    const hint = getTwoTenJackHint(makeState({ hint: { reason: 'custom' } }));
    expect(hint?.reason).toBe('custom');
  });
});
