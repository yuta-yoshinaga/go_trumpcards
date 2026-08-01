import { describe, expect, it } from 'vitest';
import type { BlackHoleResponse } from '../../types/card';
import { getBlackHoleHint } from './blackholeHint';

function makeState(overrides?: Partial<BlackHoleResponse>): BlackHoleResponse {
  return {
    fans: [[], [], []],
    blackHole: [],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getBlackHoleHint', () => {
  it('names the fan the backend recommends', () => {
    const hint = getBlackHoleHint(makeState({ hint: { fan: 4 } }));
    expect(hint?.targetAction).toBe('fan-4');
    expect(hint?.confidence).toBe('strong');
  });

  it('returns null without a hint', () => {
    expect(getBlackHoleHint(makeState())).toBeNull();
  });

  // **fan が -1 のヒントは「対象なし」。**山番号として使うと fan--1 になる。
  it('returns null for a negative fan index', () => {
    expect(getBlackHoleHint(makeState({ hint: { fan: -1 } }))).toBeNull();
  });

  it('returns null once the game has ended', () => {
    expect(getBlackHoleHint(makeState({ phase: 1, hint: { fan: 0 } }))).toBeNull();
  });
});
