import { describe, expect, it } from 'vitest';
import type { AuldLangSyneResponse } from '../../types/card';
import { getAuldLangSyneHint } from './auldlangsyneHint';

function makeState(overrides: Partial<AuldLangSyneResponse> = {}): AuldLangSyneResponse {
  return {
    foundations: [[], [], [], []],
    wastes: [[], [], [], []],
    stockCount: 0,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getAuldLangSyneHint', () => {
  it('returns null when the game has cleared', () => {
    expect(getAuldLangSyneHint(makeState({ phase: 1, hint: { wasteIdx: 0, foundationIdx: 0 } }))).toBeNull();
  });

  it('returns null when the game is over', () => {
    expect(getAuldLangSyneHint(makeState({ phase: 2, hint: { wasteIdx: 0, foundationIdx: 0 } }))).toBeNull();
  });

  it('returns null when there is no backend hint', () => {
    expect(getAuldLangSyneHint(makeState())).toBeNull();
  });

  it('returns a waste hint when the backend indicates waste → foundation', () => {
    const hint = getAuldLangSyneHint(makeState({ hint: { wasteIdx: 2, foundationIdx: 3 } }));
    expect(hint).not.toBeNull();
    expect(hint?.reason).toBe('frontendHint.auldlangsyneWaste');
    expect(hint?.targetAction).toBe('waste2-to-f3');
    expect(hint?.confidence).toBe('strong');
  });

  it('keeps the waste index distinct from the foundation index', () => {
    const hint = getAuldLangSyneHint(makeState({ hint: { wasteIdx: 0, foundationIdx: 1 } }));
    expect(hint?.targetAction).toBe('waste0-to-f1');
  });
});
