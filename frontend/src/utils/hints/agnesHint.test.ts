import { describe, expect, it } from 'vitest';
import type { AgnesResponse } from '../../types/card';
import { getAgnesHint } from './agnesHint';

function makeState(overrides: Partial<AgnesResponse> = {}): AgnesResponse {
  return {
    tableau: [[], [], [], [], [], [], []],
    stockCount: 0,
    foundation: [[], [], [], []],
    baseRank: 1,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getAgnesHint', () => {
  it('returns null when no hint in response', () => {
    expect(getAgnesHint(makeState())).toBeNull();
  });

  it('returns null when game is cleared (phase 1)', () => {
    expect(
      getAgnesHint(
        makeState({
          phase: 1,
          hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
        }),
      ),
    ).toBeNull();
  });

  it('returns foundation hint with strong confidence', () => {
    const hint = getAgnesHint(
      makeState({ hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 } }),
    );
    expect(hint).toEqual({
      targetAction: 'moveToFoundation',
      reason: 'frontendHint.moveToFoundation',
      confidence: 'strong',
    });
  });

  it('returns tableau hint with moderate confidence', () => {
    const hint = getAgnesHint(
      makeState({ hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'tableau', toCol: 2 } }),
    );
    expect(hint).toEqual({
      targetAction: 'moveToTableau',
      reason: 'frontendHint.moveToTableau',
      confidence: 'moderate',
    });
  });
});
