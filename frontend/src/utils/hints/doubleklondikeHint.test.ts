import { describe, expect, it } from 'vitest';
import type { DoubleKlondikeResponse } from '../../types/card';
import { getDoubleKlondikeHint } from './doubleklondikeHint';

function makeState(overrides?: Partial<DoubleKlondikeResponse>): DoubleKlondikeResponse {
  return {
    tableau: [],
    foundation: [],
    stockCount: 40,
    waste: [],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getDoubleKlondikeHint', () => {
  it('rates a foundation move as strong', () => {
    const hint = getDoubleKlondikeHint(
      makeState({ hint: { fromZone: 'tableau', fromCol: 3, cardIndex: 0, toZone: 'foundation', toCol: 1 } }),
    );
    expect(hint?.targetAction).toBe('tableau-3');
    expect(hint?.confidence).toBe('strong');
  });

  it('rates a tableau move as moderate', () => {
    const hint = getDoubleKlondikeHint(
      makeState({ hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 2, toZone: 'tableau', toCol: 5 } }),
    );
    expect(hint?.reason).toBe('frontendHint.dklondikeToTableau');
    expect(hint?.confidence).toBe('moderate');
  });

  // **廃札には列が無い。**-1 を付けて waste--1 にしない。
  it('names the waste without a column', () => {
    const hint = getDoubleKlondikeHint(
      makeState({ hint: { fromZone: 'waste', fromCol: -1, cardIndex: -1, toZone: 'foundation', toCol: 0 } }),
    );
    expect(hint?.targetAction).toBe('waste');
  });

  it('returns null without a hint', () => {
    expect(getDoubleKlondikeHint(makeState())).toBeNull();
  });

  it('returns null when a zone is missing', () => {
    expect(
      getDoubleKlondikeHint(
        makeState({ hint: { fromZone: '', fromCol: 0, cardIndex: 0, toZone: 'tableau', toCol: 1 } }),
      ),
    ).toBeNull();
  });

  it('returns null once the game has ended', () => {
    expect(
      getDoubleKlondikeHint(
        makeState({
          phase: 1,
          hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
        }),
      ),
    ).toBeNull();
  });
});
