import { describe, expect, it } from 'vitest';
import type { WhistConfig, WhistResponse } from '../../types/card';
import { getWhistHint } from './whistHint';

const defaultConfig: WhistConfig = { cpuDifficulty: 0, pointLimit: 5 };

function makeState(overrides: Partial<WhistResponse> = {}): WhistResponse {
  return {
    players: [
      { id: 0, isHuman: true, cards: [], cardCount: 5, trickCount: 0, roundScore: 0, cumulativeScore: 0, team: 0 },
      { id: 1, isHuman: false, cards: [], cardCount: 5, trickCount: 0, roundScore: 0, cumulativeScore: 0, team: 0 },
      { id: 2, isHuman: false, cards: [], cardCount: 5, trickCount: 0, roundScore: 0, cumulativeScore: 0, team: 0 },
      { id: 3, isHuman: false, cards: [], cardCount: 5, trickCount: 0, roundScore: 0, cumulativeScore: 0, team: 0 },
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

describe('getWhistHint', () => {
  it('returns null when game ended', () => {
    expect(getWhistHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when no hint in response', () => {
    expect(getWhistHint(makeState())).toBeNull();
  });

  it('returns null when hint has no cardIndex', () => {
    expect(getWhistHint(makeState({ hint: { reason: 'lead_strong' } }))).toBeNull();
  });

  it('maps lead_strong reason to hint.leadStrategic', () => {
    const hint = getWhistHint(makeState({ hint: { cardIndex: 2, reason: 'lead_strong' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.leadStrategic', confidence: 'strong' });
  });

  it('maps follow_suit reason to hint.followSuit', () => {
    const hint = getWhistHint(makeState({ hint: { cardIndex: 0, reason: 'follow_suit' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' });
  });

  it('maps trump_cut reason to hint.trumpCut', () => {
    const hint = getWhistHint(makeState({ hint: { cardIndex: 1, reason: 'trump_cut' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.trumpCut', confidence: 'moderate' });
  });

  it('maps discard_high reason to hint.discardLowest', () => {
    const hint = getWhistHint(makeState({ hint: { cardIndex: 3, reason: 'discard_high' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.discardLowest', confidence: 'moderate' });
  });

  it('falls back to generic reason for unknown backend reason', () => {
    const hint = getWhistHint(makeState({ hint: { cardIndex: 0, reason: 'unknown_reason' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.leadStrategic', confidence: 'moderate' });
  });
});
