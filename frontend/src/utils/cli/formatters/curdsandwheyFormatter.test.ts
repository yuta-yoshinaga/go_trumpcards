import { describe, expect, it } from 'vitest';
import type { CurdsAndWheyResponse } from '../../../types/card';
import { CurdsAndWheyPhase } from '../../../types/phases';
import { formatCurdsAndWheyState } from './curdsandwheyFormatter';

function state(overrides: Partial<CurdsAndWheyResponse> = {}): CurdsAndWheyResponse {
  return {
    columns: [
      [
        { design: 'SPADE', value: 9 },
        { design: 'SPADE', value: 8 },
      ],
      [],
      ...Array.from({ length: 11 }, () => []),
    ],
    completedSuits: 1,
    phase: CurdsAndWheyPhase.PLAYING,
    moveCount: 4,
    canUndo: true,
    legalTargets: { 'tableau-0-1': { tableau: [1], foundation: [] } },
    runStarts: [[1], ...Array.from({ length: 12 }, () => [])],
    message: '',
    ...overrides,
  };
}

describe('formatCurdsAndWheyState', () => {
  it('formats columns, zero-based card indices, and legal move guidance', () => {
    const text = formatCurdsAndWheyState(state());
    expect(text).toContain('phase: Playing');
    expect(text).toContain('Column 0: [0]♠9  [1]*♠8');
    expect(text).toContain('p 0 1 1');
    expect(text).toContain('h (hint)');
  });
  it('formats clear and game over states with reset only and suppresses server messages', () => {
    const clear = formatCurdsAndWheyState(state({ phase: CurdsAndWheyPhase.GAME_CLEAR, message: 'サーバ文言' }));
    expect(clear).toContain('Game clear!');
    expect(clear).toContain('Available command: r (reset)');
    expect(clear).not.toContain('サーバ文言');
    expect(formatCurdsAndWheyState(state({ phase: CurdsAndWheyPhase.GAME_OVER }))).toContain('No moves remain.');
  });
  it('prints a hint only for a requested hintAvailable response', () => {
    const hint = { fromCol: 0, cardIndex: 1, toCol: 1 };
    expect(formatCurdsAndWheyState(state({ hint }))).not.toContain('HINT:');
    expect(formatCurdsAndWheyState(state({ hint, messageCode: 'curdsandwhey.hintAvailable' }))).toContain(
      'HINT: Move ♠8 from column 0 to column 1.',
    );
  });
  it('prints the no-hint response in English', () => {
    expect(
      formatCurdsAndWheyState(state({ message: 'ヒントはありません', messageCode: 'curdsandwhey.noHint' })),
    ).toContain('No hint available.');
  });
});
