import { describe, expect, it } from 'vitest';
import type { RussianSolitaireResponse } from '../../types/card';
import { getRussianSolitaireHint } from './russianSolitaireHint';

function makeState(overrides: Partial<RussianSolitaireResponse> = {}): RussianSolitaireResponse {
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

describe('getRussianSolitaireHint', () => {
  it('returns null when no hint in response', () => {
    expect(getRussianSolitaireHint(makeState())).toBeNull();
  });

  it('returns foundation hint when toZone is foundation', () => {
    const hint = getRussianSolitaireHint(
      makeState({ hint: { fromCol: 2, cardIndex: 0, toZone: 'foundation', toCol: 1 } }),
    );
    expect(hint).toEqual({
      targetAction: 'moveToFoundation',
      reason: 'frontendHint.moveToFoundation',
      confidence: 'strong',
    });
  });

  it('returns tableau hint when toZone is tableau', () => {
    const hint = getRussianSolitaireHint(makeState({ hint: { fromCol: 3, cardIndex: 2, toZone: 'tableau', toCol: 5 } }));
    expect(hint).toEqual({
      targetAction: 'moveToTableau',
      reason: 'frontendHint.moveToTableau',
      confidence: 'moderate',
    });
  });

  it('returns moderate confidence for unknown toZone', () => {
    const hint = getRussianSolitaireHint(makeState({ hint: { fromCol: 0, cardIndex: 0, toZone: 'other', toCol: 0 } }));
    expect(hint).toEqual({
      targetAction: 'moveToTableau',
      reason: 'frontendHint.moveToTableau',
      confidence: 'moderate',
    });
  });

  it('returns null when game is cleared (phase 1)', () => {
    expect(
      getRussianSolitaireHint(
        makeState({ phase: 1, hint: { fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 } }),
      ),
    ).toBeNull();
  });
});
