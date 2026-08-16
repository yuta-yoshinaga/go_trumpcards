import { describe, expect, it } from 'vitest';
import type { CruelResponse } from '../../types/card';
import { getCruelHint } from './cruelHint';

function makeState(overrides: Partial<CruelResponse> = {}): CruelResponse {
  return {
    tableau: [[], [], [], [], [], [], [], [], [], [], [], []],
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canAutoComplete: false,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getCruelHint', () => {
  it('returns null when no hint and not stalemate', () => {
    expect(getCruelHint(makeState())).toBeNull();
  });

  it('returns shift suggestion in stalemate without backend hint', () => {
    const hint = getCruelHint(makeState({ isStalemate: true }));
    expect(hint).toEqual({
      targetAction: 'shift',
      reason: 'frontendHint.shiftRecommended',
      confidence: 'moderate',
    });
  });

  it('returns foundation hint when toZone is foundation', () => {
    const hint = getCruelHint(makeState({ hint: { fromCol: 2, cardIndex: 0, toZone: 'foundation', toCol: 1 } }));
    expect(hint).toEqual({
      targetAction: 'moveToFoundation',
      reason: 'frontendHint.moveToFoundation',
      confidence: 'strong',
    });
  });

  it('returns tableau hint when toZone is tableau', () => {
    const hint = getCruelHint(makeState({ hint: { fromCol: 3, cardIndex: 0, toZone: 'tableau', toCol: 5 } }));
    expect(hint).toEqual({
      targetAction: 'moveToTableau',
      reason: 'frontendHint.moveToTableau',
      confidence: 'moderate',
    });
  });

  it('returns null when game has cleared', () => {
    expect(
      getCruelHint(makeState({ phase: 1, hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 } })),
    ).toBeNull();
  });
});
