import { describe, expect, it } from 'vitest';
import type { PyramidResponse } from '../../types/card';
import { PyramidPhase } from '../../types/phases';
import { getPyramidHint } from './pyramidHint';

function makeState(overrides: Partial<PyramidResponse> = {}): PyramidResponse {
  return {
    pyramid: [
      [{ card: { design: 'SPADE', value: 5 }, removed: false, exposed: true }],
      [
        { card: { design: 'HEART', value: 8 }, removed: false, exposed: true },
        { card: { design: 'DIAMOND', value: 3 }, removed: false, exposed: true },
      ],
    ],
    stockCount: 10,
    waste: [],
    phase: PyramidPhase.PLAYING,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getPyramidHint', () => {
  it('returns null when game is cleared', () => {
    expect(getPyramidHint(makeState({ phase: PyramidPhase.GAME_CLEAR }))).toBeNull();
  });

  it('returns null when game is over', () => {
    expect(getPyramidHint(makeState({ phase: PyramidPhase.GAME_OVER }))).toBeNull();
  });

  it('suggests removing King when exposed', () => {
    const state = makeState({
      pyramid: [[{ card: { design: 'SPADE', value: 13 }, removed: false, exposed: true }]],
    });
    const hint = getPyramidHint(state);
    expect(hint?.targetAction).toBe('remove');
    expect(hint?.reason).toBe('frontendHint.removeKing');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests removing pair summing to 13', () => {
    const state = makeState({
      pyramid: [
        [{ card: { design: 'SPADE', value: 6 }, removed: false, exposed: true }],
        [{ card: { design: 'HEART', value: 7 }, removed: false, exposed: true }],
      ],
    });
    const hint = getPyramidHint(state);
    expect(hint?.targetAction).toBe('remove');
    expect(hint?.reason).toBe('frontendHint.removePair');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests using waste King', () => {
    const state = makeState({
      pyramid: [[{ card: { design: 'SPADE', value: 2 }, removed: false, exposed: true }]],
      waste: [{ design: 'CLOVER', value: 13 }],
    });
    const hint = getPyramidHint(state);
    expect(hint?.reason).toBe('frontendHint.useWasteKing');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests using waste + exposed pair', () => {
    const state = makeState({
      pyramid: [[{ card: { design: 'SPADE', value: 4 }, removed: false, exposed: true }]],
      waste: [{ design: 'HEART', value: 9 }],
    });
    const hint = getPyramidHint(state);
    expect(hint?.reason).toBe('frontendHint.useWaste');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests drawing from stock as fallback', () => {
    const state = makeState({
      pyramid: [
        [{ card: { design: 'SPADE', value: 2 }, removed: false, exposed: true }],
        [{ card: { design: 'HEART', value: 3 }, removed: false, exposed: true }],
      ],
      stockCount: 5,
      waste: [],
    });
    const hint = getPyramidHint(state);
    expect(hint?.reason).toBe('frontendHint.drawFromStock');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns null when no moves and no stock', () => {
    const state = makeState({
      pyramid: [
        [{ card: { design: 'SPADE', value: 2 }, removed: false, exposed: true }],
        [{ card: { design: 'HEART', value: 3 }, removed: false, exposed: true }],
      ],
      stockCount: 0,
      waste: [],
    });
    expect(getPyramidHint(state)).toBeNull();
  });

  it('ignores removed cards', () => {
    const state = makeState({
      pyramid: [
        [{ card: { design: 'SPADE', value: 13 }, removed: true, exposed: true }],
        [{ card: { design: 'HEART', value: 2 }, removed: false, exposed: true }],
      ],
      stockCount: 0,
      waste: [],
    });
    expect(getPyramidHint(state)).toBeNull();
  });

  it('ignores non-exposed cards', () => {
    const state = makeState({
      pyramid: [
        [{ card: { design: 'SPADE', value: 13 }, removed: false, exposed: false }],
        [{ card: { design: 'HEART', value: 2 }, removed: false, exposed: true }],
      ],
      stockCount: 0,
      waste: [],
    });
    expect(getPyramidHint(state)).toBeNull();
  });
});
