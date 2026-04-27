import { describe, expect, it } from 'vitest';
import type { SlapjackResponse } from '../../types/card';
import { getSlapjackHint } from './slapjackHint';

function baseState(overrides: Partial<SlapjackResponse> = {}): SlapjackResponse {
  return {
    phase: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    currentTurnIdx: 0,
    isHumanTurn: true,
    isTopJack: false,
    centerPileSize: 0,
    topCard: null,
    players: [
      { name: 'You', isHuman: true, stockSize: 26 },
      { name: 'CPU', isHuman: false, stockSize: 26 },
    ],
    cpuDifficulty: 1,
    pendingKind: 0,
    pendingDeadlineMs: 0,
    lastEventKind: 0,
    lastEventPlayerIdx: 0,
    message: '',
    ...overrides,
  };
}

describe('getSlapjackHint', () => {
  it('returns null when game ended', () => {
    expect(getSlapjackHint(baseState({ gameEndFlag: true }))).toBeNull();
  });

  it('suggests slap when J is on top', () => {
    const hint = getSlapjackHint(baseState({ isTopJack: true, centerPileSize: 1 }));
    expect(hint?.targetAction).toBe('slap');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests step on human turn (no Jack)', () => {
    const hint = getSlapjackHint(baseState({ isHumanTurn: true }));
    expect(hint?.targetAction).toBe('step');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns null on CPU turn (no Jack)', () => {
    expect(getSlapjackHint(baseState({ isHumanTurn: false }))).toBeNull();
  });
});
