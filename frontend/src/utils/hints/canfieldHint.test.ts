import { describe, expect, it } from 'vitest';
import type { CanfieldResponse } from '../../types/card';
import { getCanfieldHint } from './canfieldHint';

function makeState(overrides: Partial<CanfieldResponse> = {}): CanfieldResponse {
  return {
    tableau: [[], [], [], []],
    reserve: [],
    stockCount: 0,
    waste: [],
    foundation: [[], [], [], []],
    baseRank: 1,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

describe('getCanfieldHint', () => {
  it('returns null when no hint in response', () => {
    expect(getCanfieldHint(makeState())).toBeNull();
  });

  it('returns null when game is cleared (phase 1)', () => {
    expect(
      getCanfieldHint(
        makeState({
          phase: 1,
          hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
        }),
      ),
    ).toBeNull();
  });

  it('returns foundation hint with strong confidence', () => {
    const hint = getCanfieldHint(
      makeState({ hint: { fromZone: 'reserve', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 } }),
    );
    expect(hint).toEqual({
      targetAction: 'moveToFoundation',
      reason: 'frontendHint.moveToFoundation',
      confidence: 'strong',
    });
  });

  it('returns tableau hint with moderate confidence', () => {
    const hint = getCanfieldHint(
      makeState({ hint: { fromZone: 'waste', fromCol: 0, cardIndex: 0, toZone: 'tableau', toCol: 2 } }),
    );
    expect(hint).toEqual({
      targetAction: 'moveToTableau',
      reason: 'frontendHint.moveToTableau',
      confidence: 'moderate',
    });
  });
});
