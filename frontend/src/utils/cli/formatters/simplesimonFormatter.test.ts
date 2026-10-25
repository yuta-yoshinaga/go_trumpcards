import { describe, expect, it } from 'vitest';
import type { SimpleSimonResponse } from '../../../types/card';
import { SimpleSimonPhase } from '../../../types/phases';
import { formatSimpleSimonState } from './simplesimonFormatter';

const base = {
  columns: [
    [
      { design: 'SPADE', value: 3 },
      { design: 'HEART', value: 4 },
    ],
    [],
  ],
  runStarts: [[1], []],
  legalTargets: {},
  phase: SimpleSimonPhase.PLAYING,
  moveCount: 4,
  canUndo: false,
  completedSuits: 1,
  message: '',
} as unknown as SimpleSimonResponse;
describe('formatSimpleSimonState', () => {
  it('shows indexed cards, movable marker, phase and available actions', () => {
    const out = formatSimpleSimonState(base);
    expect(out).toContain('phase: Playing');
    expect(out).toContain('C0 (bottom → top):  ♠3[0] *♥4[1]');
    expect(out).toContain('m <fromCol>');
  });
  it('formats requested move hints and suppresses passive hints', () => {
    const hinted = {
      ...base,
      hint: { fromCol: 0, cardIndex: 1, toCol: 4 },
      messageCode: 'simplesimon.hintAvailable',
    } as SimpleSimonResponse;
    expect(formatSimpleSimonState(hinted)).toContain('HINT: move [0] [1] to column 4.');
    expect(formatSimpleSimonState({ ...hinted, messageCode: undefined })).not.toContain('HINT');
    expect(formatSimpleSimonState({ ...base, messageCode: 'simplesimon.noHint' })).toContain('No hint available.');
  });
  it('shows only reset after the game ends', () => {
    expect(formatSimpleSimonState({ ...base, phase: SimpleSimonPhase.GAME_CLEAR })).toContain('Commands: r');
    expect(formatSimpleSimonState({ ...base, phase: SimpleSimonPhase.GAME_CLEAR, canUndo: true })).toContain(
      'Commands: r, u',
    );
    expect(formatSimpleSimonState({ ...base, canUndo: true })).toContain(
      'Commands: m <fromCol> <cardIndex> <toCol>, h, g, u, r',
    );
  });
});
