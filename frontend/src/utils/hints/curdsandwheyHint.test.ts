import { describe, expect, it } from 'vitest';
import type { CurdsAndWheyResponse } from '../../types/card';
import { getCurdsAndWheyHint } from './curdsandwheyHint';

function makeState(overrides?: Partial<CurdsAndWheyResponse>): CurdsAndWheyResponse {
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

describe('getCurdsAndWheyHint', () => {
  it('names the source column the backend recommends', () => {
    const hint = getCurdsAndWheyHint(makeState({ hint: { fromCol: 3, cardIndex: 0, toCol: 5 } }));
    expect(hint?.targetAction).toBe('col-3');
    expect(hint?.confidence).toBe('moderate');
  });

  it('returns null without a hint', () => {
    expect(getCurdsAndWheyHint(makeState())).toBeNull();
  });

  // **負の列番号は「対象なし」。**そのまま使うと col--1 になる。
  it('returns null for a negative column', () => {
    expect(getCurdsAndWheyHint(makeState({ hint: { fromCol: -1, cardIndex: 0, toCol: 2 } }))).toBeNull();
    expect(getCurdsAndWheyHint(makeState({ hint: { fromCol: 2, cardIndex: 0, toCol: -1 } }))).toBeNull();
  });

  it('returns null once the game has ended', () => {
    expect(getCurdsAndWheyHint(makeState({ phase: 1, hint: { fromCol: 0, cardIndex: 0, toCol: 1 } }))).toBeNull();
  });
});
