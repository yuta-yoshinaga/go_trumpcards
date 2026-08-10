import { describe, expect, it } from 'vitest';
import type { OsmosisResponse } from '../../types/card';
import { getOsmosisHint } from './osmosisHint';

function makeState(overrides: Partial<OsmosisResponse> = {}): OsmosisResponse {
  return {
    reserve: [[], [], [], []],
    stockCount: 0,
    waste: [],
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

describe('getOsmosisHint', () => {
  it('returns null when no hint in response', () => {
    expect(getOsmosisHint(makeState())).toBeNull();
  });

  it('returns null when game is cleared (phase 1)', () => {
    expect(
      getOsmosisHint(
        makeState({
          phase: 1,
          hint: { fromZone: 'reserve', fromCol: 0, toCol: 0 },
        }),
      ),
    ).toBeNull();
  });

  it('returns foundation hint with strong confidence', () => {
    const hint = getOsmosisHint(makeState({ hint: { fromZone: 'waste', fromCol: -1, toCol: 1 } }));
    expect(hint).toEqual({
      targetAction: 'moveToFoundation',
      reason: 'frontendHint.moveToFoundation',
      confidence: 'strong',
    });
  });
});
