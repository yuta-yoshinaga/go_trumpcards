import { describe, expect, it } from 'vitest';
import type { SimpleSimonResponse } from '../../types/card';
import { getSimpleSimonHint } from './simplesimonHint';

function makeState(overrides?: Partial<SimpleSimonResponse>): SimpleSimonResponse {
  return {
    columns: [[], [], []],
    completedSuits: 0,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

describe('getSimpleSimonHint', () => {
  it('names the source column the backend recommends', () => {
    const hint = getSimpleSimonHint(makeState({ hint: { fromCol: 3, cardIndex: 0, toCol: 5 } }));
    expect(hint?.targetAction).toBe('col-3');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns null without a hint', () => {
    expect(getSimpleSimonHint(makeState())).toBeNull();
  });

  // **負の列番号は「対象なし」。**そのまま使うと col--1 になる。
  it('returns null for a negative column', () => {
    expect(getSimpleSimonHint(makeState({ hint: { fromCol: -1, cardIndex: 0, toCol: 2 } }))).toBeNull();
    expect(getSimpleSimonHint(makeState({ hint: { fromCol: 2, cardIndex: 0, toCol: -1 } }))).toBeNull();
  });

  it('returns null once the game has ended', () => {
    expect(getSimpleSimonHint(makeState({ phase: 1, hint: { fromCol: 0, cardIndex: 0, toCol: 1 } }))).toBeNull();
  });
});
