import { describe, expect, it } from 'vitest';
import type { EgyptianRatscrewResponse } from '../../types/card';
import { getEgyptianRatscrewHint } from './egyptianratscrewHint';

function baseState(overrides: Partial<EgyptianRatscrewResponse> = {}): EgyptianRatscrewResponse {
  return {
    phase: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    currentTurnIdx: 0,
    isHumanTurn: true,
    isTopFaceCard: false,
    isSlappable: false,
    centerPileSize: 0,
    topCard: null,
    players: [
      { name: 'You', isHuman: true, stockSize: 26 },
      { name: 'CPU', isHuman: false, stockSize: 26 },
    ],
    cpuDifficulty: 1,
    chanceRemaining: 0,
    faceChances: { jack: 1, queen: 2, king: 3, ace: 4 },
    chanceFromIdx: -1,
    pendingKind: 0,
    pendingDeadlineMs: 0,
    lastEventKind: 0,
    lastEventPlayerIdx: 0,
    lastSlapReason: 0,
    message: '',
    ...overrides,
  };
}

describe('getEgyptianRatscrewHint', () => {
  it('returns null when game ended', () => {
    expect(getEgyptianRatscrewHint(baseState({ gameEndFlag: true }))).toBeNull();
  });

  it('suggests slap when pile is slappable', () => {
    const hint = getEgyptianRatscrewHint(baseState({ isSlappable: true, centerPileSize: 2 }));
    expect(hint?.targetAction).toBe('slap');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests step on human turn (not slappable)', () => {
    const hint = getEgyptianRatscrewHint(baseState({ isHumanTurn: true }));
    expect(hint?.targetAction).toBe('step');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns null on CPU turn (not slappable)', () => {
    expect(getEgyptianRatscrewHint(baseState({ isHumanTurn: false }))).toBeNull();
  });
});
