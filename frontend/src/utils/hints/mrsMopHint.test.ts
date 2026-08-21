import { describe, expect, it } from 'vitest';
import type { Card, MrsMopResponse, MrsMopTableauCard } from '../../types/card';
import { MrsMopPhase } from '../../types/phases';
import { getMrsMopHint } from './mrsMopHint';

const S: Card['design'] = 'SPADE';
const H: Card['design'] = 'HEART';
const D: Card['design'] = 'DIAMOND';
const C: Card['design'] = 'CLOVER';

function tc(design: Card['design'], value: number, faceUp = true): MrsMopTableauCard {
  return { card: { design, value }, faceUp };
}

function makeState(overrides: Partial<MrsMopResponse> = {}): MrsMopResponse {
  return {
    tableau: [[tc(S, 10), tc(S, 9), tc(S, 8)], [tc(H, 5, false), tc(H, 7), tc(H, 6)], [], [], [], [], [], [], [], []],
    stockCount: 0,
    completedSuits: 0,
    phase: MrsMopPhase.PLAYING,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    score: 500,
    difficulty: 1,
    message: '',
    ...overrides,
  };
}

describe('getMrsMopHint', () => {
  it('returns null when game is cleared', () => {
    expect(getMrsMopHint(makeState({ phase: MrsMopPhase.GAME_CLEAR }))).toBeNull();
  });

  it('returns null when game is over', () => {
    expect(getMrsMopHint(makeState({ phase: MrsMopPhase.GAME_OVER }))).toBeNull();
  });

  it('suggests completing suit for near-complete sequence', () => {
    // Build a 10-card same-suit sequence
    const longSeq: MrsMopTableauCard[] = [];
    for (let v = 13; v >= 4; v--) {
      longSeq.push(tc(S, v));
    }
    const state = makeState({
      tableau: [longSeq, [tc(H, 5)], [], [], [], [], [], [], [], []],
    });
    const hint = getMrsMopHint(state);
    expect(hint?.reason).toBe('frontendHint.completeSuit');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests same-suit build move', () => {
    const state = makeState({
      tableau: [[tc(S, 8)], [tc(S, 9)], [], [], [], [], [], [], [], []],
      stockCount: 0,
    });
    const hint = getMrsMopHint(state);
    expect(hint?.reason).toBe('frontendHint.buildSameSuit');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests revealing face-down cards', () => {
    const state = makeState({
      tableau: [[tc(S, 5, false), tc(H, 4)], [tc(D, 10)], [], [], [], [], [], [], [], []],
      stockCount: 0,
    });
    const hint = getMrsMopHint(state);
    expect(hint?.reason).toBe('frontendHint.revealFaceDown');
    expect(hint?.confidence).toBe('moderate');
  });

  it('suggests using empty column', () => {
    const state = makeState({
      tableau: [[tc(S, 5)], [tc(H, 10)], [], [], [], [], [], [], [], []],
      stockCount: 0,
    });
    const hint = getMrsMopHint(state);
    expect(hint?.reason).toBe('frontendHint.useEmptyColumn');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns null when no moves and no stock', () => {
    const state = makeState({
      tableau: [
        [tc(S, 1)],
        [tc(H, 1)],
        [tc(D, 1)],
        [tc(C, 1)],
        [tc(S, 13)],
        [tc(H, 13)],
        [tc(D, 13)],
        [tc(C, 13)],
        [tc(S, 7)],
        [tc(H, 7)],
      ],
      stockCount: 0,
    });
    expect(getMrsMopHint(state)).toBeNull();
  });
});
