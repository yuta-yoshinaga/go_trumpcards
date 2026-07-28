import { describe, expect, it } from 'vitest';
import type { SirTommyResponse } from '../../types/card';
import { getSirTommyHint } from './sirtommyHint';

function makeState(overrides: Partial<SirTommyResponse> = {}): SirTommyResponse {
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

describe('getSirTommyHint', () => {
  it('returns null when the game has cleared', () => {
    expect(
      getSirTommyHint(makeState({ phase: 1, hint: { fromZone: 'stock', wasteIdx: -1, foundationIdx: 0 } })),
    ).toBeNull();
  });

  it('returns null when the game is over', () => {
    expect(
      getSirTommyHint(makeState({ phase: 2, hint: { fromZone: 'stock', wasteIdx: -1, foundationIdx: 0 } })),
    ).toBeNull();
  });

  it('returns null when there is no backend hint', () => {
    expect(getSirTommyHint(makeState())).toBeNull();
  });

  it('returns a stock hint when the backend indicates stock → foundation', () => {
    const hint = getSirTommyHint(makeState({ hint: { fromZone: 'stock', wasteIdx: -1, foundationIdx: 2 } }));
    expect(hint).not.toBeNull();
    expect(hint?.reason).toBe('frontendHint.sirtommyStock');
    expect(hint?.targetAction).toBe('stock-to-f2');
    expect(hint?.confidence).toBe('strong');
  });

  it('returns a waste hint when the backend indicates waste → foundation', () => {
    const hint = getSirTommyHint(makeState({ hint: { fromZone: 'waste', wasteIdx: 1, foundationIdx: 0 } }));
    expect(hint).not.toBeNull();
    expect(hint?.reason).toBe('frontendHint.sirtommyWaste');
    expect(hint?.targetAction).toBe('waste1-to-f0');
    expect(hint?.confidence).toBe('strong');
  });
});
