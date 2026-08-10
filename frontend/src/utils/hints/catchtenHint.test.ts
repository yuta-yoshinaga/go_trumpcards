import { describe, expect, it } from 'vitest';
import type { CatchTenConfig, CatchTenResponse } from '../../types/card';
import { getCatchTenHint } from './catchtenHint';

const defaultConfig: CatchTenConfig = { cpuDifficulty: 0, pointLimit: 41 };

function makeState(overrides: Partial<CatchTenResponse> = {}): CatchTenResponse {
  return {
    players: [
      { id: 0, isHuman: true, cards: [], cardCount: 9, trickCount: 0, roundScore: 0, cumulativeScore: 0, team: 0 },
      { id: 1, isHuman: false, cards: [], cardCount: 9, trickCount: 0, roundScore: 0, cumulativeScore: 0, team: 1 },
      { id: 2, isHuman: false, cards: [], cardCount: 9, trickCount: 0, roundScore: 0, cumulativeScore: 0, team: 0 },
      { id: 3, isHuman: false, cards: [], cardCount: 9, trickCount: 0, roundScore: 0, cumulativeScore: 0, team: 1 },
    ],
    phase: 0,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 3,
    dealerIdx: 3,
    teamScores: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    validPlayIndices: [],
    message: '',
    config: defaultConfig,
    ...overrides,
  };
}

describe('getCatchTenHint', () => {
  it('returns null when game ended', () => {
    expect(getCatchTenHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when no hint in response', () => {
    expect(getCatchTenHint(makeState())).toBeNull();
  });

  it('returns null when hint has no cardIndex', () => {
    expect(getCatchTenHint(makeState({ hint: { reason: 'lead_strong' } }))).toBeNull();
  });

  it('maps lead_strong reason to hint.leadStrategic', () => {
    const hint = getCatchTenHint(makeState({ hint: { cardIndex: 2, reason: 'lead_strong' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.leadStrategic', confidence: 'strong' });
  });

  it('maps follow_suit reason to hint.followSuit', () => {
    const hint = getCatchTenHint(makeState({ hint: { cardIndex: 0, reason: 'follow_suit' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' });
  });

  it('maps trump_cut reason to hint.trumpCut', () => {
    const hint = getCatchTenHint(makeState({ hint: { cardIndex: 1, reason: 'trump_cut' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.trumpCut', confidence: 'moderate' });
  });

  it('maps discard_high reason to hint.discardLowest', () => {
    const hint = getCatchTenHint(makeState({ hint: { cardIndex: 3, reason: 'discard_high' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.discardLowest', confidence: 'moderate' });
  });

  it('falls back to generic reason for unknown backend reason', () => {
    const hint = getCatchTenHint(makeState({ hint: { cardIndex: 0, reason: 'unknown_reason' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.leadStrategic', confidence: 'moderate' });
  });
});
