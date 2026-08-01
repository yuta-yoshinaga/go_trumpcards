import { describe, expect, it } from 'vitest';
import type { PyramidResponse } from '../../../types/card';
import { formatPyramidState } from './pyramidFormatter';

function makeState(overrides?: Partial<PyramidResponse>): PyramidResponse {
  return {
    pyramid: [
      [{ card: { design: 'SPADE', value: 1 }, removed: false, exposed: false }],
      [
        { card: { design: 'HEART', value: 6 }, removed: false, exposed: true },
        { card: { design: 'CLOVER', value: 7 }, removed: true, exposed: true },
      ],
    ],
    stockCount: 20,
    waste: [{ design: 'DIAMOND', value: 4 }],
    phase: 0,
    moveCount: 8,
    canUndo: true,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatPyramidState', () => {
  it('formats the basic state', () => {
    const output = formatPyramidState(makeState());
    expect(output).toContain('Pyramid');
    expect(output).toContain('stock: 20');
    expect(output).toContain('moves: 8');
  });

  it('hides a card that is not exposed', () => {
    expect(formatPyramidState(makeState())).toContain('??');
  });

  it('reports a stalemate', () => {
    expect(formatPyramidState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { type: 'pair', row1: 1, col1: 0, row2: 1, col2: 1 };
    expect(formatPyramidState(makeState({ hint, messageCode: 'pyramid.hintAvailable' }))).toContain('HINT: (1,0)');
    expect(formatPyramidState(makeState({ hint, messageCode: 'pyramid.playing' }))).not.toContain('HINT:');
  });
});
