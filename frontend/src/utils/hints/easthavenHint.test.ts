import { describe, expect, it } from 'vitest';
import type { EasthavenResponse } from '../../types/card';
import { getEasthavenHint } from './easthavenHint';

function makeState(overrides: Partial<EasthavenResponse> = {}): EasthavenResponse {
  return {
    tableau: [[], [], [], [], [], [], []],
    foundation: [[], [], [], []],
    stockCount: 31,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getEasthavenHint', () => {
  it('returns null when no hint in response', () => {
    expect(getEasthavenHint(makeState())).toBeNull();
  });

  it('returns foundation hint when toZone is foundation', () => {
    const hint = getEasthavenHint(makeState({ hint: { fromCol: 2, cardIndex: 0, toZone: 'foundation', toCol: 1 } }));
    expect(hint).toEqual({
      targetAction: 'moveToFoundation',
      reason: 'frontendHint.moveToFoundation',
      confidence: 'strong',
    });
  });

  it('returns tableau hint when toZone is tableau', () => {
    const hint = getEasthavenHint(makeState({ hint: { fromCol: 3, cardIndex: 2, toZone: 'tableau', toCol: 5 } }));
    expect(hint).toEqual({
      targetAction: 'moveToTableau',
      reason: 'frontendHint.moveToTableau',
      confidence: 'moderate',
    });
  });

  it('returns null when game is cleared (phase 1)', () => {
    expect(
      getEasthavenHint(makeState({ phase: 1, hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 } })),
    ).toBeNull();
  });
});
