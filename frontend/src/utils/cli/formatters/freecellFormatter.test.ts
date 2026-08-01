import { describe, expect, it } from 'vitest';
import type { FreeCellResponse } from '../../../types/card';
import { formatFreecellState } from './freecellFormatter';

function makeState(overrides?: Partial<FreeCellResponse>): FreeCellResponse {
  return {
    tableau: [[{ design: 'SPADE', value: 13 }], [], [], [], [], [], [], []],
    freeCells: [null, null, null, null],
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 3,
    canUndo: true,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatFreecellState', () => {
  it('formats the basic state', () => {
    const output = formatFreecellState(makeState());
    expect(output).toContain('FreeCell');
    expect(output).toContain('cells:');
    expect(output).toContain('foundation:');
    expect(output).toContain('moves: 3');
    expect(output).toContain('col0:');
  });

  it('marks empty columns', () => {
    expect(formatFreecellState(makeState())).toContain('col1: [empty]');
  });

  it('reports a stalemate', () => {
    expect(formatFreecellState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 };
    expect(formatFreecellState(makeState({ hint, messageCode: 'freecell.hintAvailable' }))).toContain('HINT:');
    expect(formatFreecellState(makeState({ hint, messageCode: 'freecell.playing' }))).not.toContain('HINT:');
  });
});
