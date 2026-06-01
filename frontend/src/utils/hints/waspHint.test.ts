import { describe, expect, it } from 'vitest';
import type { WaspResponse } from '../../types/card';
import { getWaspHint } from './waspHint';

function makeState(overrides: Partial<WaspResponse> = {}): WaspResponse {
  return {
    tableau: [],
    stockCount: 0,
    completedSuits: 0,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getWaspHint', () => {
  it('returns null when the game is over (clear)', () => {
    const state = makeState({ phase: 1, hint: { fromCol: 0, cardIndex: 0, toCol: 1 } });
    expect(getWaspHint(state)).toBeNull();
  });

  it('returns null when phase is game over (2)', () => {
    const state = makeState({ phase: 2, hint: { fromCol: 0, cardIndex: 0, toCol: 1 } });
    expect(getWaspHint(state)).toBeNull();
  });

  it('returns null when there is no backend hint', () => {
    expect(getWaspHint(makeState())).toBeNull();
  });

  it('returns a moveToTableau hint for a tableau move', () => {
    const state = makeState({ hint: { fromCol: 0, cardIndex: 2, toCol: 3 } });
    const hint = getWaspHint(state);
    expect(hint).not.toBeNull();
    expect(hint?.reason).toBe('frontendHint.moveToTableau');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns a dealStock hint when backend suggests deal (fromCol < 0)', () => {
    const state = makeState({ hint: { fromCol: -1, cardIndex: -1, toCol: -1 } });
    const hint = getWaspHint(state);
    expect(hint).not.toBeNull();
    expect(hint?.reason).toBe('frontendHint.dealStock');
  });
});
