import { describe, expect, it } from 'vitest';
import type { TriPeaksResponse } from '../../types/card';
import { TriPeaksPhase } from '../../types/phases';
import { getTriPeaksHint } from './tripeaksHint';

function makeState(overrides: Partial<TriPeaksResponse> = {}): TriPeaksResponse {
  return {
    layout: [
      [{ card: { design: 'SPADE', value: 5 }, removed: false, exposed: true }],
      [{ card: { design: 'HEART', value: 8 }, removed: false, exposed: true }],
    ],
    stockCount: 10,
    waste: [{ design: 'DIAMOND', value: 4 }],
    phase: TriPeaksPhase.PLAYING,
    moveCount: 0,
    score: 0,
    combo: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getTriPeaksHint', () => {
  it('returns null when game is cleared', () => {
    expect(getTriPeaksHint(makeState({ phase: TriPeaksPhase.GAME_CLEAR }))).toBeNull();
  });

  it('returns null when game is over', () => {
    expect(getTriPeaksHint(makeState({ phase: TriPeaksPhase.GAME_OVER }))).toBeNull();
  });

  it('suggests drawing when no waste card', () => {
    const hint = getTriPeaksHint(makeState({ waste: [], stockCount: 5 }));
    expect(hint?.reason).toBe('frontendHint.drawFromStock');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns null when no waste and no stock', () => {
    expect(getTriPeaksHint(makeState({ waste: [], stockCount: 0 }))).toBeNull();
  });

  it('suggests removing single adjacent card', () => {
    const state = makeState({
      layout: [[{ card: { design: 'SPADE', value: 5 }, removed: false, exposed: true }]],
      waste: [{ design: 'HEART', value: 4 }],
    });
    const hint = getTriPeaksHint(state);
    expect(hint?.reason).toBe('frontendHint.canRemove');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests multiple removable cards', () => {
    const state = makeState({
      layout: [
        [{ card: { design: 'SPADE', value: 5 }, removed: false, exposed: true }],
        [{ card: { design: 'HEART', value: 3 }, removed: false, exposed: true }],
      ],
      waste: [{ design: 'DIAMOND', value: 4 }],
    });
    const hint = getTriPeaksHint(state);
    expect(hint?.reason).toBe('frontendHint.multipleRemovable');
    expect(hint?.confidence).toBe('strong');
  });

  it('handles King-Ace wrap-around', () => {
    const state = makeState({
      layout: [[{ card: { design: 'SPADE', value: 1 }, removed: false, exposed: true }]],
      waste: [{ design: 'HEART', value: 13 }],
    });
    const hint = getTriPeaksHint(state);
    expect(hint?.reason).toBe('frontendHint.canRemove');
  });

  it('suggests drawing when no adjacent cards', () => {
    const state = makeState({
      layout: [[{ card: { design: 'SPADE', value: 10 }, removed: false, exposed: true }]],
      waste: [{ design: 'HEART', value: 4 }],
      stockCount: 5,
    });
    const hint = getTriPeaksHint(state);
    expect(hint?.reason).toBe('frontendHint.drawFromStock');
  });

  it('returns null when no moves and no stock', () => {
    const state = makeState({
      layout: [[{ card: { design: 'SPADE', value: 10 }, removed: false, exposed: true }]],
      waste: [{ design: 'HEART', value: 4 }],
      stockCount: 0,
    });
    expect(getTriPeaksHint(state)).toBeNull();
  });

  it('ignores removed cards', () => {
    const state = makeState({
      layout: [[{ card: { design: 'SPADE', value: 5 }, removed: true, exposed: true }]],
      waste: [{ design: 'HEART', value: 4 }],
      stockCount: 0,
    });
    expect(getTriPeaksHint(state)).toBeNull();
  });
});
