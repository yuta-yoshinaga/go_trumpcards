import { describe, expect, it } from 'vitest';
import type { YukonResponse } from '../../types/card';
import { getYukonHint } from './yukonHint';

function makeState(overrides: Partial<YukonResponse> = {}): YukonResponse {
  return {
    tableau: [[], [], [], [], [], [], []],
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getYukonHint', () => {
  it('returns null when no hint in response', () => {
    expect(getYukonHint(makeState())).toBeNull();
  });

  it('returns foundation hint when toZone is foundation', () => {
    const hint = getYukonHint(makeState({ hint: { fromCol: 2, cardIndex: 0, toZone: 'foundation', toCol: 1 } }));
    expect(hint).toEqual({
      targetAction: 'moveToFoundation',
      reason: 'frontendHint.moveToFoundation',
      confidence: 'strong',
    });
  });

  it('returns tableau hint when toZone is tableau', () => {
    const hint = getYukonHint(makeState({ hint: { fromCol: 3, cardIndex: 2, toZone: 'tableau', toCol: 5 } }));
    expect(hint).toEqual({
      targetAction: 'moveToTableau',
      reason: 'frontendHint.moveToTableau',
      confidence: 'moderate',
    });
  });

  it('returns moderate confidence for unknown toZone', () => {
    const hint = getYukonHint(makeState({ hint: { fromCol: 0, cardIndex: 0, toZone: 'other', toCol: 0 } }));
    expect(hint).toEqual({
      targetAction: 'moveToTableau',
      reason: 'frontendHint.moveToTableau',
      confidence: 'moderate',
    });
  });

  it('returns null when game is cleared (phase 1)', () => {
    expect(
      getYukonHint(makeState({ phase: 1, hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 } })),
    ).toBeNull();
  });
});
