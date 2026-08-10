import { describe, expect, it } from 'vitest';
import type { ColoradoResponse } from '../../types/card';
import { getColoradoHint } from './coloradoHint';

function makeState(overrides: Partial<ColoradoResponse> = {}): ColoradoResponse {
  return {
    tableau: [],
    foundation: [],
    foundationAscending: [true, true, true, true, false, false, false, false],
    stockCount: 71,
    waste: [],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

describe('getColoradoHint', () => {
  it('returns null when the game has cleared', () => {
    expect(
      getColoradoHint(
        makeState({ phase: 1, hint: { fromZone: 'waste', fromIdx: -1, toZone: 'foundation', toIdx: 0 } }),
      ),
    ).toBeNull();
  });

  it('returns null when the backend sent no hint', () => {
    expect(getColoradoHint(makeState())).toBeNull();
  });

  it('points at a waste-to-foundation move', () => {
    expect(
      getColoradoHint(makeState({ hint: { fromZone: 'waste', fromIdx: -1, toZone: 'foundation', toIdx: 5 } })),
    ).toEqual({ targetAction: 'waste-to-f5', reason: 'frontendHint.coloradoWaste', confidence: 'strong' });
  });

  it('points at a tableau-to-foundation move', () => {
    expect(
      getColoradoHint(makeState({ hint: { fromZone: 'tableau', fromIdx: 12, toZone: 'foundation', toIdx: 0 } })),
    ).toEqual({ targetAction: 't12-to-f0', reason: 'frontendHint.coloradoTableau', confidence: 'strong' });
  });

  it('points at filling a gap from the stock', () => {
    expect(
      getColoradoHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 3 } })),
    ).toEqual({ targetAction: 'stock-to-t3', reason: 'frontendHint.coloradoFillGap', confidence: 'moderate' });
  });

  it('points at drawing when that is all the backend offers', () => {
    expect(
      getColoradoHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 } })),
    ).toEqual({ targetAction: 'draw', reason: 'frontendHint.coloradoDraw', confidence: 'moderate' });
  });

  // "Bury the waste somewhere" is always legal once the stock is gone, so
  // highlighting it would light up a pile on nearly every turn for no reason.
  it('stays silent on the last-resort bury move', () => {
    expect(
      getColoradoHint(makeState({ hint: { fromZone: 'waste', fromIdx: -1, toZone: 'tableau', toIdx: 8 } })),
    ).toBeNull();
  });
});
