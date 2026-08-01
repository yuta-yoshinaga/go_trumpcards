import { describe, expect, it } from 'vitest';
import type { TriPeaksResponse } from '../../../types/card';
import { formatTripeaksState } from './tripeaksFormatter';

function makeState(overrides?: Partial<TriPeaksResponse>): TriPeaksResponse {
  return {
    layout: [
      [{ card: { design: 'SPADE', value: 1 }, removed: false, exposed: false }],
      [
        { card: { design: 'HEART', value: 6 }, removed: false, exposed: true },
        { card: { design: 'CLOVER', value: 7 }, removed: true, exposed: true },
      ],
    ],
    stockCount: 18,
    waste: [{ design: 'DIAMOND', value: 5 }],
    phase: 0,
    moveCount: 4,
    canUndo: true,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatTripeaksState', () => {
  it('formats the basic state', () => {
    const output = formatTripeaksState(makeState());
    expect(output).toContain('TriPeaks');
    expect(output).toContain('stock: 18');
    expect(output).toContain('moves: 4');
  });

  it('hides a card that is not exposed', () => {
    expect(formatTripeaksState(makeState())).toContain('??');
  });

  it('reports a stalemate', () => {
    expect(formatTripeaksState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { type: 'remove', row: 1, col: 0 };
    expect(formatTripeaksState(makeState({ hint, messageCode: 'tripeaks.hintAvailable' }))).toContain('HINT: (1,0)');
    expect(formatTripeaksState(makeState({ hint, messageCode: 'tripeaks.playing' }))).not.toContain('HINT:');
  });
});
