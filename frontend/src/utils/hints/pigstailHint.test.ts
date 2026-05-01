import { describe, expect, it } from 'vitest';
import type { PigsTailResponse } from '../../types/card';
import { getPigstailHint } from './pigstailHint';

function makeState(overrides?: Partial<PigsTailResponse>): PigsTailResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 0, cards: [] },
      { id: 1, isHuman: false, cardCount: 0, cards: [] },
    ],
    circleCount: 0,
    centerTop: null,
    centerCount: 52,
    currentTurn: 0,
    gameEndFlag: false,
    loserIdx: -1,
    lastDrawCard: null,
    lastPenalty: false,
    cpuActions: [],
    humanAction: null,
    message: '',
    ...overrides,
  };
}

describe('getPigstailHint', () => {
  it('returns draw hint when human turn and no penalty', () => {
    const result = getPigstailHint(makeState());
    expect(result).toEqual({ targetAction: 'draw', reason: 'hint.draw', confidence: 'moderate' });
  });

  it('returns penalty-aware hint after penalty', () => {
    const result = getPigstailHint(makeState({ lastPenalty: true }));
    expect(result?.reason).toBe('hint.afterPenalty');
  });

  it('returns null when not human turn', () => {
    expect(getPigstailHint(makeState({ currentTurn: 1 }))).toBeNull();
  });

  it('returns null when game ended', () => {
    expect(getPigstailHint(makeState({ gameEndFlag: true }))).toBeNull();
  });
});
