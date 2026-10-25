import { describe, expect, it } from 'vitest';
import type { CurdsAndWheyResponse } from '../../../types/card';
import { CurdsAndWheyPhase } from '../../../types/phases';
import { parseCurdsAndWheyCommand } from './curdsandwheyCommands';

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
    completedSuits: 0,
    phase: CurdsAndWheyPhase.PLAYING,
    moveCount: 0,
    canUndo: true,
    legalTargets: {},
    // Array order is bottom to top; the descending same-suit run starts at 0.
    runStarts: [[0, 1], ...Array.from({ length: 12 }, () => [])],
    message: '',
    ...overrides,
  };
}

describe('parseCurdsAndWheyCommand', () => {
  const s = state();
  it.each([
    ['p 0 1 1', ['m', { fromCol: 0, cardIndex: 1, toCol: 1 }]],
    ['play 0 1 2', ['m', { fromCol: 0, cardIndex: 1, toCol: 2 }]],
    ['u', ['u']],
    ['undo', ['u']],
    ['un 2', ['undo_n', { n: 2 }]],
    ['undo_n 3', ['undo_n', { n: 3 }]],
    ['g', ['g']],
    ['giveup', ['g']],
    ['h', ['hint']],
    ['hint', ['hint']],
    ['r', ['reset']],
    ['reset', ['reset']],
  ] as const)('maps %s to the matching page action', (input, expected) => {
    expect(parseCurdsAndWheyCommand(input, s)).toEqual({ args: expected });
  });

  it.each(['p', 'p 0 1', 'p x 1 2'])('rejects missing or invalid move arguments: %s', (input) => {
    expect(parseCurdsAndWheyCommand(input, s)).toHaveProperty('error');
  });
  it.each(['p -1 1 2', 'p 13 1 2', 'p 0 2 2', 'p 0 1 13', 'p 0 1 0'])(
    'rejects out of range or illegal move: %s',
    (input) => {
      expect(parseCurdsAndWheyCommand(input, s)).toHaveProperty('error');
    },
  );
  it.each(['un', 'un 0', 'un x'])('rejects invalid undo counts: %s', (input) => {
    expect(parseCurdsAndWheyCommand(input, s)).toHaveProperty('error');
  });
  it('rejects undo_n when undo is unavailable', () => {
    expect(parseCurdsAndWheyCommand('un 1', state({ canUndo: false }))).toHaveProperty('error');
  });
  it('allows moving a partial run, rejects a buried card, and treats the last index as the top', () => {
    const columns = Array.from({ length: 13 }, () => [] as { design: 'SPADE' | 'HEART'; value: number }[]);
    columns[3] = [
      { design: 'SPADE', value: 9 },
      { design: 'HEART', value: 6 },
      { design: 'HEART', value: 5 },
      { design: 'HEART', value: 4 },
    ];
    const stacked = state({
      columns,
      // The immovable bottom card is buried; each later index begins a movable suffix.
      runStarts: Array.from({ length: 13 }, (_, col) => (col === 3 ? [1, 2, 3] : [])),
    });
    expect(parseCurdsAndWheyCommand('p 3 2 4', stacked)).toEqual({
      args: ['m', { fromCol: 3, cardIndex: 2, toCol: 4 }],
    });
    expect(parseCurdsAndWheyCommand('p 3 0 4', stacked)).toHaveProperty('error');
    expect(stacked.columns[3]?.at(-1)).toEqual({ design: 'HEART', value: 4 });
    expect(stacked.runStarts[3]).toContain(3);
  });
  it('rejects unavailable undo and actions outside play', () => {
    expect(parseCurdsAndWheyCommand('u', state({ canUndo: false }))).toHaveProperty('error');
    const ended = state({ phase: CurdsAndWheyPhase.GAME_OVER });
    for (const input of ['p 0 1 1', 'g', 'h']) expect(parseCurdsAndWheyCommand(input, ended)).toHaveProperty('error');
  });
  it('rejects unknown commands', () => expect(parseCurdsAndWheyCommand('wat', s)).toHaveProperty('error'));
});
