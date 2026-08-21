import { describe, expect, it } from 'vitest';
import type { StalactitesResponse } from '../../../types/card';
import { formatStalactitesState } from './stalactitesFormatter';

function makeState(overrides?: Partial<StalactitesResponse>): StalactitesResponse {
  return {
    tableau: [[{ design: 'SPADE', value: 13 }], [], [], [], [], [], [], []],
    baseRank: 1,
    cells: [null, null, null, null],
    foundation: [[], [], [], []],
    maxMovableCards: 320,
    maxMovableCardsToEmptyColumn: 160,
    phase: 0,
    moveCount: 3,
    canUndo: true,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatStalactitesState', () => {
  it('formats the basic state', () => {
    const output = formatStalactitesState(makeState());
    expect(output).toContain('Stalactites');
    expect(output).toContain('cells:');
    expect(output).toContain('foundation:');
    expect(output).toContain('moves: 3');
    expect(output).toContain('col0:');
  });

  it('marks empty columns', () => {
    expect(formatStalactitesState(makeState())).toContain('col1: [empty]');
  });

  it('reports a stalemate', () => {
    expect(formatStalactitesState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 };
    expect(formatStalactitesState(makeState({ hint, messageCode: 'stalactites.hintAvailable' }))).toContain('HINT:');
    expect(formatStalactitesState(makeState({ hint, messageCode: 'stalactites.playing' }))).not.toContain('HINT:');
  });
});
