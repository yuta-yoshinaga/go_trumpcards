import { describe, expect, it } from 'vitest';
import type { FourteenOutResponse } from '../types/card';
import {
  fourteenOutCanUndo,
  fourteenOutPartners,
  fourteenOutRemovablePairs,
  fourteenOutTail,
} from './fourteenOutRules';

const state: FourteenOutResponse = {
  columns: [
    [{ card: { design: 'CLOVER', value: 2 } }, { card: { design: 'SPADE', value: 9 } }],
    [{ card: { design: 'HEART', value: 5 } }],
    [],
  ],
  phase: 0,
  removedCount: 0,
  removablePairs: 1,
  legalTargets: { 'tableau-0': { tableau: [1], foundation: [] }, 'tableau-1': { tableau: [0], foundation: [] } },
  canUndo: true,
  isStalemate: false,
  message: '',
  messageCode: 'fourteenout.playing',
};

describe('Fourteen Out rules', () => {
  it('reads the final card as the top and treats missing or empty columns as no card', () => {
    expect(fourteenOutTail(state.columns[0])?.value).toBe(9);
    expect(fourteenOutTail(state.columns[2])).toBeNull();
    expect(fourteenOutTail(undefined)).toBeNull();
  });

  it('uses server targets and exposes undo availability', () => {
    expect(fourteenOutPartners(state, 0)).toEqual([1]);
    expect(fourteenOutPartners(state, 2)).toEqual([]);
    expect(fourteenOutCanUndo(state)).toBe(true);
    expect(fourteenOutCanUndo({ ...state, canUndo: false })).toBe(false);
  });

  it('deduplicates and sorts pairs by numeric column indices', () => {
    const many = {
      ...state,
      columns: Array.from({ length: 12 }, () => []),
      legalTargets: {
        'tableau-10': { tableau: [11], foundation: [] },
        'tableau-2': { tableau: [10, 3], foundation: [] },
        'tableau-3': { tableau: [2], foundation: [] },
        'tableau-11': { tableau: [10], foundation: [] },
      },
    } as FourteenOutResponse;
    expect(fourteenOutRemovablePairs(many)).toEqual([
      [2, 3],
      [2, 10],
      [10, 11],
    ]);
  });
});
