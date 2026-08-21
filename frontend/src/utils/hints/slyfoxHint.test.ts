import { describe, expect, it } from 'vitest';
import type { SlyFoxResponse } from '../../types/card';
import { getSlyFoxHint } from './slyfoxHint';

function makeState(overrides: Partial<SlyFoxResponse> = {}): SlyFoxResponse {
  return {
    tableau: [],
    foundation: [],
    foundationAscending: [true, true, true, true, false, false, false, false],
    stockCount: 71,
    dealtThisCycle: 20,
    dealCycle: 20,
    reserveLocked: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

describe('getSlyFoxHint', () => {
  it('returns null when the game has cleared', () => {
    expect(
      getSlyFoxHint(makeState({ phase: 1, hint: { fromZone: 'waste', fromIdx: -1, toZone: 'foundation', toIdx: 0 } })),
    ).toBeNull();
  });

  it('returns null when the backend sent no hint', () => {
    expect(getSlyFoxHint(makeState())).toBeNull();
  });

  it('points at a reserve-to-foundation move', () => {
    expect(
      getSlyFoxHint(makeState({ hint: { fromZone: 'tableau', fromIdx: 12, toZone: 'foundation', toIdx: 0 } })),
    ).toEqual({ targetAction: 't12-to-f0', reason: 'frontendHint.slyFoxTableau', confidence: 'strong' });
  });

  // **組札へ直接送れる札は 20 枚に数えない**ので、常にこちらが得。枠へ配る手と
  // 同じ扱いにすると、その差が読み手に伝わらない。
  it('prefers dealing straight to a foundation', () => {
    expect(
      getSlyFoxHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'foundation', toIdx: 5 } })),
    ).toEqual({
      targetAction: 'deal-to-f5',
      reason: 'frontendHint.slyFoxDealToFoundation',
      confidence: 'strong',
    });
  });

  it('points at the least costly slot to deal onto', () => {
    expect(getSlyFoxHint(makeState({ hint: { fromZone: 'stock', fromIdx: -1, toZone: 'tableau', toIdx: 3 } }))).toEqual(
      { targetAction: 'deal-to-t3', reason: 'frontendHint.slyFoxDealToSlot', confidence: 'moderate' },
    );
  });

  // 負のコントロール: リザーブから枠への手はこのゲームに無い。届いても黙る。
  it('stays silent on a reserve-to-reserve hint, which the game does not have', () => {
    expect(
      getSlyFoxHint(makeState({ hint: { fromZone: 'tableau', fromIdx: 1, toZone: 'tableau', toIdx: 8 } })),
    ).toBeNull();
  });
});
