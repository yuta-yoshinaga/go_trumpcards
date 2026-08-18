import { describe, expect, it } from 'vitest';
import type { BristolResponse } from '../../types/card';
import { getBristolHint } from './bristolHint';

function makeState(overrides: Partial<BristolResponse> = {}): BristolResponse {
  return {
    tableau: [[], [], [], [], [], [], [], []],
    fan: [[], [], []],
    stockCount: 0,
    foundation: [[], [], [], []],
    legalTargets: {},
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    undoToEscape: 0,
    message: '',
    ...overrides,
  };
}

describe('getBristolHint', () => {
  it('returns null when no hint in response', () => {
    expect(getBristolHint(makeState())).toBeNull();
  });

  it('returns null when game is cleared (phase 1)', () => {
    expect(
      getBristolHint(
        makeState({
          phase: 1,
          hint: { fromZone: 'tableau', fromCol: 0, toZone: 'foundation', toCol: 0 },
        }),
      ),
    ).toBeNull();
  });

  it('returns foundation hint with strong confidence', () => {
    const hint = getBristolHint(
      makeState({ hint: { fromZone: 'tableau', fromCol: 0, toZone: 'foundation', toCol: 1 } }),
    );
    expect(hint).toEqual({
      targetAction: 'moveToFoundation',
      reason: 'frontendHint.moveToFoundation',
      confidence: 'strong',
    });
  });

  it('returns tableau hint with moderate confidence', () => {
    const hint = getBristolHint(makeState({ hint: { fromZone: 'fan', fromCol: 1, toZone: 'tableau', toCol: 3 } }));
    expect(hint).toEqual({
      targetAction: 'moveToTableau',
      reason: 'frontendHint.moveToTableau',
      confidence: 'moderate',
    });
  });
});
