import { describe, expect, it } from 'vitest';
import type { SalicLawResponse } from '../types/card';
import { isSalicLawTarget, listSalicLawTargets } from './saliclawTargets';

function makeState(overrides?: Partial<SalicLawResponse>): SalicLawResponse {
  return {
    tableau: Array.from({ length: 8 }, () => []),
    foundation: Array.from({ length: 8 }, () => []),
    stockCount: 95,
    queens: [],
    openPiles: 2,
    canAutoComplete: false,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

const card = (value: number) => ({ design: 'SPADE' as const, value });

describe('isSalicLawTarget', () => {
  it('allows an ace on an open empty foundation and ignores suit for the next rank', () => {
    const tableau = [[], [card(13), card(1)]];
    expect(isSalicLawTarget(makeState({ tableau }), 1, 'foundation', 0)).toBe(true);
    expect(
      isSalicLawTarget(makeState({ tableau: [[], [card(13), card(2)]], foundation: [[card(1)]] }), 1, 'foundation', 0),
    ).toBe(true);
  });

  it('rejects illegal foundation targets and ranks', () => {
    const tableau = [[], [card(13), card(2)]];
    expect(isSalicLawTarget(makeState({ tableau }), 1, 'foundation', 0)).toBe(false);
    expect(isSalicLawTarget(makeState({ tableau: [[], [card(13), card(1)]] }), 1, 'foundation', 2)).toBe(false);
    expect(isSalicLawTarget(makeState({ tableau, openPiles: 1 }), 1, 'foundation', 1)).toBe(false);
    expect(isSalicLawTarget(makeState({ tableau }), 1, 'foundation', -1)).toBe(false);
    expect(isSalicLawTarget(makeState({ tableau }), 1, 'foundation', 8)).toBe(false);
    expect(isSalicLawTarget(makeState({ tableau, foundation: [] }), 1, 'foundation', 0)).toBe(false);
    const missingTop: { value: number }[] = [];
    missingTop.length = 1;
    expect(
      isSalicLawTarget(
        makeState({ tableau, foundation: [missingTop] as SalicLawResponse['foundation'] }),
        1,
        'foundation',
        0,
      ),
    ).toBe(false);
    expect(
      isSalicLawTarget(
        makeState({
          tableau: [[], [card(13), card(12)]],
          foundation: [Array.from({ length: 11 }, (_, i) => card(i + 1))],
        }),
        1,
        'foundation',
        0,
      ),
    ).toBe(false);
  });

  it('allows only a different tableau column containing exactly its king', () => {
    const state = makeState({ tableau: [[card(13)], [card(13), card(7)]] });
    expect(isSalicLawTarget(state, 1, 'tableau', 0)).toBe(true);
    expect(isSalicLawTarget(state, 1, 'tableau', 1)).toBe(false);
    expect(isSalicLawTarget(state, 1, 'tableau', 2)).toBe(false);
  });

  it('rejects a missing, empty, single-card, or valueless source', () => {
    expect(isSalicLawTarget(makeState(), 9, 'foundation', 0)).toBe(false);
    expect(isSalicLawTarget(makeState({ tableau: [[card(13)]] }), 0, 'foundation', 0)).toBe(false);
    expect(isSalicLawTarget(makeState({ tableau: [[card(13), undefined as never]] }), 0, 'foundation', 0)).toBe(false);
    expect(isSalicLawTarget(makeState({ tableau: [[card(13), card(1)]] }), 9, 'foundation', 0)).toBe(false);
  });
});

describe('listSalicLawTargets', () => {
  it('lists legal destinations across foundations and bare king columns', () => {
    const state = makeState({
      tableau: [[card(13)], [card(13), card(1)], [card(13)], [], [], [], [], []],
    });
    const targets = listSalicLawTargets(state, 1);
    expect(targets).toHaveLength(4);
    expect(targets).toEqual([
      { zone: 'foundation', idx: 0 },
      { zone: 'foundation', idx: 1 },
      { zone: 'tableau', idx: 0 },
      { zone: 'tableau', idx: 2 },
    ]);
  });

  it('returns zero when the selection is absent or has no legal destination', () => {
    expect(listSalicLawTargets(makeState(), 0)).toEqual([]);
    expect(listSalicLawTargets(makeState({ tableau: [[card(13), card(4)]] }), 0)).toEqual([]);
  });
});
