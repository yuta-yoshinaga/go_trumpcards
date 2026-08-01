import { describe, expect, it } from 'vitest';
import type { GolfResponse } from '../../../types/card';
import { formatGolfState } from './golfFormatter';

function makeState(overrides?: Partial<GolfResponse>): GolfResponse {
  return {
    layout: [
      [{ card: { design: 'SPADE', value: 13 }, removed: false, exposed: true }],
      [{ card: { design: 'HEART', value: 7 }, removed: false, exposed: false }],
    ],
    stockCount: 12,
    waste: [{ design: 'CLOVER', value: 5 }],
    phase: 0,
    moveCount: 6,
    canUndo: true,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('formatGolfState', () => {
  it('formats the basic state', () => {
    const output = formatGolfState(makeState());
    expect(output).toContain('Golf');
    expect(output).toContain('stock: 12');
    expect(output).toContain('moves: 6');
    expect(output).toContain('[0]');
  });

  it('hides a card that is not exposed', () => {
    expect(formatGolfState(makeState())).toContain('??');
  });

  it('reports a stalemate', () => {
    expect(formatGolfState(makeState({ isStalemate: true }))).toContain('Stalemate');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { type: 'remove', col: 1 };
    expect(formatGolfState(makeState({ hint, messageCode: 'golf.hintAvailable' }))).toContain('HINT: col 1');
    expect(formatGolfState(makeState({ hint, messageCode: 'golf.playing' }))).not.toContain('HINT:');
  });
});
