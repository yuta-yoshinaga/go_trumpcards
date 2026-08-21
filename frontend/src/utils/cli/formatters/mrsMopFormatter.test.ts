import { describe, expect, it } from 'vitest';
import type { MrsMopResponse } from '../../../types/card';
import { formatMrsMopState } from './mrsMopFormatter';

function makeState(overrides?: Partial<MrsMopResponse>): MrsMopResponse {
  return {
    tableau: [[{ card: { design: 'SPADE', value: 13 }, faceUp: true }], []],
    stockCount: 50,
    completedSuits: 1,
    phase: 0,
    moveCount: 7,
    canUndo: true,
    isStalemate: false,
    score: 480,
    difficulty: 1,
    message: '',
    ...overrides,
  };
}

describe('formatMrsMopState', () => {
  it('formats the basic state', () => {
    const output = formatMrsMopState(makeState());
    expect(output).toContain('MrsMop Solitaire');
    expect(output).toContain('stock: 50');
    expect(output).toContain('completed: 1/8');
    expect(output).toContain('moves: 7');
  });

  it('names the difficulty', () => {
    expect(formatMrsMopState(makeState())).toContain('1 suit');
    expect(formatMrsMopState(makeState({ difficulty: 4 }))).toContain('4 suits');
  });

  it('marks empty columns and face-down cards', () => {
    const output = formatMrsMopState(
      makeState({
        tableau: [[{ card: { design: 'HEART', value: 2 }, faceUp: false }], []],
      }),
    );
    expect(output).toContain('[?]');
    expect(output).toContain('col1: [empty]');
  });

  // **HINT 行は hint を頼んだときだけ。**受動ヒントが Output に載るように
  // なった (#4483) ので、messageCode で「頼んだ応答か」を見分ける。
  it('shows the hint only when the hint was requested', () => {
    const hint = { fromCol: 0, cardIndex: 0, toCol: 1 };
    expect(formatMrsMopState(makeState({ hint, messageCode: 'mrsmop.hintAvailable' }))).toContain('HINT:');
    expect(formatMrsMopState(makeState({ hint, messageCode: 'mrsmop.playing' }))).not.toContain('HINT:');
  });
});
