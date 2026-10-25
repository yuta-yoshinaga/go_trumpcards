import { describe, expect, it } from 'vitest';
import type { FourteenOutResponse } from '../../../types/card';
import { FourteenOutPhase } from '../../../types/phases';
import { formatFourteenOutState } from './fourteenoutFormatter';

const playing: FourteenOutResponse = {
  columns: [
    [{ card: { design: 'SPADE', value: 3 } }, { card: { design: 'SPADE', value: 9 } }],
    [{ card: { design: 'HEART', value: 5 } }],
    [],
    ...Array.from({ length: 9 }, () => []),
  ],
  phase: FourteenOutPhase.PLAYING,
  removedCount: 0,
  removablePairs: 1,
  legalTargets: { 'tableau-0': { tableau: [1], foundation: [] } },
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'fourteenout.playing',
};

describe('formatFourteenOutState', () => {
  it('shows indexed columns, playable top cards, and actionable commands', () => {
    const text = formatFourteenOutState(playing);
    expect(text).toContain('phase: Playing');
    expect(text).toContain('[0] x♠3 *♠9');
    expect(text).toContain('[1] *♥5');
    expect(text).toContain('[2] [empty]');
    expect(text).toContain('p 0 1');
    expect(text).toContain('h (hint)');
    expect(text).toContain('g (give up)');
  });

  it('shows requested hint and server no-hint response only', () => {
    const passive = formatFourteenOutState({ ...playing, hint: { action: 'remove', fromCol: 0, toCol: 1 } });
    expect(passive).not.toContain('HINT:');
    const requested = formatFourteenOutState({
      ...playing,
      hint: { action: 'remove', fromCol: 0, toCol: 1 },
      messageCode: 'fourteenout.hintAvailable',
    });
    expect(requested).toContain('HINT: Remove columns 0 and 1.');
    expect(formatFourteenOutState({ ...playing, messageCode: 'fourteenout.noHint' })).toContain('No hint available.');
    expect(formatFourteenOutState({ ...playing, hint: { action: 'remove', fromCol: 0, toCol: 1 } })).not.toContain(
      'No hint',
    );
  });

  it('shows terminal phase commands and English outcome text', () => {
    const clear = formatFourteenOutState({
      ...playing,
      phase: FourteenOutPhase.GAME_CLEAR,
      removedCount: 52,
      message: 'ゲームクリア！',
    });
    expect(clear).toContain('phase: Game Clear');
    expect(clear).toContain('All cards removed. Game clear!');
    expect(clear).toContain('Available command: r (reset)');
    expect(clear).not.toContain('ゲームクリア');
    const over = formatFourteenOutState({ ...playing, phase: FourteenOutPhase.GAME_OVER });
    expect(over).toContain('phase: Game Over');
    expect(over).toContain('No more moves. Game over.');
  });
});
