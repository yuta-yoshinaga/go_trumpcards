import { describe, expect, it } from 'vitest';
import type { GapsResponse } from '../../types/card';
import { GapsPhase } from '../../types/phases';
import { getGapsHint } from './gapsHint';

function baseState(overrides: Partial<GapsResponse> = {}): GapsResponse {
  return {
    grid: [[], [], [], []],
    redealsUsed: 0,
    redealsRemaining: 3,
    phase: GapsPhase.PLAYING,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getGapsHint', () => {
  it('returns null when no backend hint is present', () => {
    expect(getGapsHint(baseState())).toBeNull();
  });

  it('returns null when game is not in playing phase', () => {
    expect(
      getGapsHint(
        baseState({
          phase: GapsPhase.GAME_CLEAR,
          hint: { fromRow: 0, fromCol: 0, toRow: 0, toCol: 1 },
        }),
      ),
    ).toBeNull();
  });

  it('returns a move hint when the backend supplies one', () => {
    const result = getGapsHint(baseState({ hint: { fromRow: 1, fromCol: 0, toRow: 0, toCol: 0 } }));
    expect(result).toEqual({
      targetAction: 'move',
      reason: 'frontendHint.validMove',
      confidence: 'strong',
    });
  });
});
