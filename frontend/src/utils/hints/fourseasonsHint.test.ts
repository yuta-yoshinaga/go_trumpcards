import { describe, expect, it } from 'vitest';
import type { FourSeasonsResponse } from '../../types/card';
import { getFourSeasonsHint } from './fourseasonsHint';

function makeState(overrides: Partial<FourSeasonsResponse> = {}): FourSeasonsResponse {
  return {
    tableau: [[], [], [], [], []],
    foundation: [[], [], [], []],
    stockCount: 46,
    waste: [],
    baseRank: 7,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

describe('getFourSeasonsHint', () => {
  it('returns null when the game has cleared', () => {
    expect(
      getFourSeasonsHint(
        makeState({ phase: 1, hint: { fromZone: 'waste', fromIdx: -1, toZone: 'foundation', toIdx: 0 } }),
      ),
    ).toBeNull();
  });

  it('returns null when the game is over', () => {
    expect(
      getFourSeasonsHint(
        makeState({ phase: 2, hint: { fromZone: 'waste', fromIdx: -1, toZone: 'foundation', toIdx: 0 } }),
      ),
    ).toBeNull();
  });

  it('returns null when there is no backend hint', () => {
    expect(getFourSeasonsHint(makeState())).toBeNull();
  });

  it('reports a waste hint', () => {
    const hint = getFourSeasonsHint(
      makeState({ hint: { fromZone: 'waste', fromIdx: -1, toZone: 'foundation', toIdx: 2 } }),
    );
    expect(hint).toEqual({
      targetAction: 'waste-to-f2',
      reason: 'frontendHint.fourseasonsWaste',
      confidence: 'strong',
    });
  });

  it('reports a cross-pile hint, keeping the two indices distinct', () => {
    const hint = getFourSeasonsHint(
      makeState({ hint: { fromZone: 'tableau', fromIdx: 3, toZone: 'foundation', toIdx: 1 } }),
    );
    expect(hint).toEqual({
      targetAction: 't3-to-f1',
      reason: 'frontendHint.fourseasonsTableau',
      confidence: 'strong',
    });
  });
});
