import { describe, expect, it } from 'vitest';
import type { Card, SpideretteResponse, SpideretteTableauCard } from '../../types/card';
import { SpiderettePhase } from '../../types/phases';
import { getSpideretteHint } from './spideretteHint';

const S: Card['design'] = 'SPADE';
const H: Card['design'] = 'HEART';
const D: Card['design'] = 'DIAMOND';
const C: Card['design'] = 'CLOVER';

function tc(design: Card['design'], value: number, faceUp = true): SpideretteTableauCard {
  return { card: { design, value }, faceUp };
}

function makeState(overrides: Partial<SpideretteResponse> = {}): SpideretteResponse {
  return {
    tableau: [[tc(S, 10), tc(S, 9), tc(S, 8)], [tc(H, 5, false), tc(H, 7), tc(H, 6)], [], [], [], [], []],
    stockCount: 20,
    completedSuits: 0,
    phase: SpiderettePhase.PLAYING,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    score: 500,
    scoring: { start: 500, movePenalty: 1, suitBonus: 100 },
    message: '',
    ...overrides,
  };
}

describe('getSpideretteHint', () => {
  it('returns null for null/undefined state', () => {
    expect(getSpideretteHint(null)).toBeNull();
    expect(getSpideretteHint(undefined)).toBeNull();
  });

  it('returns null when game is cleared', () => {
    expect(getSpideretteHint(makeState({ phase: SpiderettePhase.GAME_CLEAR }))).toBeNull();
  });

  it('returns null when game is over', () => {
    expect(getSpideretteHint(makeState({ phase: SpiderettePhase.GAME_OVER }))).toBeNull();
  });

  it('suggests completing suit for near-complete sequence', () => {
    // Build a 10-card same-suit descending sequence (K→4).
    const longSeq: SpideretteTableauCard[] = [];
    for (let v = 13; v >= 4; v--) {
      longSeq.push(tc(S, v));
    }
    const state = makeState({
      tableau: [longSeq, [tc(H, 5)], [], [], [], [], []],
    });
    const hint = getSpideretteHint(state);
    expect(hint?.reason).toBe('frontendHint.completeSuit');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests same-suit build move', () => {
    const state = makeState({
      tableau: [[tc(S, 8)], [tc(S, 9)], [], [], [], [], []],
      stockCount: 0,
    });
    const hint = getSpideretteHint(state);
    expect(hint?.reason).toBe('frontendHint.buildSameSuit');
    expect(hint?.confidence).toBe('strong');
  });

  it('suggests revealing face-down cards', () => {
    const state = makeState({
      tableau: [[tc(S, 5, false), tc(H, 4)], [tc(D, 10)], [], [], [], [], []],
      stockCount: 0,
    });
    const hint = getSpideretteHint(state);
    expect(hint?.reason).toBe('frontendHint.revealFaceDown');
    expect(hint?.confidence).toBe('moderate');
  });

  it('suggests using empty column', () => {
    const state = makeState({
      tableau: [[tc(S, 5)], [tc(H, 10)], [], [], [], [], []],
      stockCount: 0,
    });
    const hint = getSpideretteHint(state);
    expect(hint?.reason).toBe('frontendHint.useEmptyColumn');
    expect(hint?.confidence).toBe('moderate');
  });

  it('suggests dealing from stock as fallback', () => {
    // Values chosen so no same-suit adjacency exists and every column is filled.
    const state = makeState({
      tableau: [[tc(S, 1)], [tc(H, 1)], [tc(D, 1)], [tc(C, 1)], [tc(S, 13)], [tc(H, 13)], [tc(D, 13)]],
      stockCount: 10,
    });
    const hint = getSpideretteHint(state);
    expect(hint?.reason).toBe('frontendHint.dealFromStock');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns null when no moves and no stock', () => {
    const state = makeState({
      tableau: [[tc(S, 1)], [tc(H, 1)], [tc(D, 1)], [tc(C, 1)], [tc(S, 13)], [tc(H, 13)], [tc(D, 13)]],
      stockCount: 0,
    });
    expect(getSpideretteHint(state)).toBeNull();
  });
});
