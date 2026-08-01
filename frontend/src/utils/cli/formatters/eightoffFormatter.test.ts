import { describe, expect, it } from 'vitest';
import type { EightOffResponse } from '../../../types/card';
import { formatEightoffState } from './eightoffFormatter';

function makeState(overrides?: Partial<EightOffResponse>): EightOffResponse {
  return {
    tableau: [[{ design: 'SPADE', value: 13 }], [], [], [], [], [], [], []],
    freeCells: [null, null, null, null, null, null, null, null],
    foundation: [[], [], [], []],
    phase: 0,
    moveCount: 4,
    canUndo: true,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatEightoffState', () => {
  it('formats the basic state', () => {
    const output = formatEightoffState(makeState());
    expect(output).toContain('Eight Off');
    expect(output).toContain('cells:');
    expect(output).toContain('moves: 4');
    expect(output).toContain('col0:');
  });

  it('marks empty columns', () => {
    expect(formatEightoffState(makeState())).toContain('col1: [empty]');
  });

  it('reports a stalemate', () => {
    expect(formatEightoffState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 1 };
    expect(formatEightoffState(makeState({ hint, messageCode: 'eightoff.hintAvailable' }))).toContain('HINT:');
    expect(formatEightoffState(makeState({ hint, messageCode: 'eightoff.playing' }))).not.toContain('HINT:');
  });
});
