import { describe, expect, it } from 'vitest';
import type { LaBelleLucieResponse } from '../../../types/card';
import { formatLaBelleLucieState } from './labellelucieFormatter';

const base: LaBelleLucieResponse = {
  fans: [[{ design: 'SPADE', value: 9 }], [{ design: 'HEART', value: 1 }]],
  movableFans: [true, false],
  foundation: [[], [], [], []],
  redealsLeft: 2,
  phase: 0,
  moveCount: 4,
  canUndo: true,
  message: '',
};

describe('formatLaBelleLucieState', () => {
  it('formats a playing state with fan cards and available actions', () => {
    const output = formatLaBelleLucieState(base);
    expect(output).toContain('phase: Playing | moves: 4 | redeals: 1/3');
    expect(output).toContain('0: ♠9*');
    expect(output).not.toContain('[0]♠9');
    expect(output).toContain('1: ♥A');
    expect(output).toContain('p <from> <to>');
  });

  it('shows requested move hints from the actual hint message code', () => {
    const output = formatLaBelleLucieState({
      ...base,
      hint: { fromFan: 1, toFan: 3, toFoundation: false },
      messageCode: 'labellelucie.hintAvailable',
    });
    expect(output).toContain('HINT: move fan 1 to fan 3');
  });

  it('shows foundation hints and suppresses passive hints without requested message codes', () => {
    expect(
      formatLaBelleLucieState({
        ...base,
        hint: { fromFan: 2, toFan: -1, toFoundation: true },
        messageCode: 'labellelucie.hintAvailable',
      }),
    ).toContain('HINT: move fan 2 to foundation');
    expect(
      formatLaBelleLucieState({
        ...base,
        hint: { fromFan: 2, toFan: -1, toFoundation: true },
        messageCode: 'labellelucie.playing',
      }),
    ).not.toContain('HINT:');
    expect(formatLaBelleLucieState({ ...base, hint: { fromFan: 2, toFan: -1, toFoundation: true } })).not.toContain(
      'HINT:',
    );
  });

  it('formats a finished game state', () => {
    const output = formatLaBelleLucieState({
      ...base,
      phase: 1,
      foundation: [[{ design: 'SPADE', value: 13 }], [], [], []],
      message: 'Cleared',
    });
    expect(output).toContain('phase: Game Clear');
    expect(output).toContain('Congratulations! You win!');
    expect(output).toContain('Cleared');
    expect(output).toContain('Game finished. Use r');
    expect(formatLaBelleLucieState({ ...base, phase: 2 })).toContain('Game over.');
  });

  it('omits redeal guidance when no redeals remain', () => {
    expect(formatLaBelleLucieState({ ...base, redealsLeft: 0 })).toContain('Other actions: ac, u, h, g');
  });

  it('marks the top (last) card of a movable fan, not the bottom one', () => {
    const output = formatLaBelleLucieState({
      ...base,
      fans: [
        [
          { design: 'SPADE', value: 9 },
          { design: 'CLOVER', value: 4 },
        ],
        [{ design: 'HEART', value: 1 }],
      ],
      movableFans: [true, false],
    });
    expect(output).toContain('0: ♠9  ♣4*');
    expect(output).not.toContain('♠9*');
  });

  it('omits undo when unavailable and reports an unavailable hint', () => {
    const unavailableUndo = formatLaBelleLucieState({ ...base, canUndo: false });
    expect(unavailableUndo).toContain('Other actions: rd, ac, h, g');
    expect(formatLaBelleLucieState({ ...base, messageCode: 'labellelucie.noHint' })).toContain('No hint available.');
  });
});
