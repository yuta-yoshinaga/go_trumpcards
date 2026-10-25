import { describe, expect, it } from 'vitest';
import type { SimpleSimonResponse } from '../../../types/card';
import { SimpleSimonPhase } from '../../../types/phases';
import { parseSimpleSimonCommand } from './simplesimonCommands';

const state = {
  columns: [[{ design: 'SPADE', value: 3 }], []],
  runStarts: [[0], []],
  legalTargets: { 'tableau-0-0': { tableau: [1], foundation: [] } },
  phase: SimpleSimonPhase.PLAYING,
  moveCount: 2,
  canUndo: true,
} as unknown as SimpleSimonResponse;
describe('parseSimpleSimonCommand', () => {
  it('maps available commands to the page API calls', () => {
    expect(parseSimpleSimonCommand('m 0 0 1', state)).toEqual({ args: ['m', { fromCol: 0, cardIndex: 0, toCol: 1 }] });
    for (const [input, args] of [
      ['g', ['g']],
      ['u', ['u']],
      ['undo_n 2', ['undo_n', { n: 2 }]],
      ['h', ['hint']],
      ['hint', ['hint']],
      ['r', ['reset']],
      ['reset', ['reset']],
    ] as const)
      expect(parseSimpleSimonCommand(input, state)).toEqual({ args });
  });
  it('allows moving a whole run to an empty column even when legalTargets omits it', () => {
    const runState = {
      ...state,
      columns: [
        [
          { design: 'SPADE', value: 5 },
          { design: 'HEART', value: 4 },
        ],
        [],
      ],
      runStarts: [[0, 1], []],
      legalTargets: { 'tableau-0-0': { tableau: [], foundation: [] } },
    } as unknown as SimpleSimonResponse;
    expect(parseSimpleSimonCommand('m 0 0 1', runState)).toEqual({
      args: ['m', { fromCol: 0, cardIndex: 0, toCol: 1 }],
    });
  });
  it('rejects unavailable moves and phase-specific commands', () => {
    expect(parseSimpleSimonCommand('m 0 1 1', state)).toHaveProperty('error');
    expect(parseSimpleSimonCommand('u', { ...state, canUndo: false })).toHaveProperty('error');
    for (const phase of [SimpleSimonPhase.GAME_CLEAR, SimpleSimonPhase.GAME_OVER]) {
      const ended = { ...state, phase };
      expect(parseSimpleSimonCommand('h', ended)).toHaveProperty('error');
      expect(parseSimpleSimonCommand('g', ended)).toHaveProperty('error');
    }
    expect(parseSimpleSimonCommand('m 0 0 1')).toHaveProperty('error');
  });
  it('rejects incomplete, out of range and illegal inputs', () => {
    for (const input of ['m', 'm 0 0', 'm -1 0 1', 'm 0 4 1', 'm 0 0 9', 'm 1 0 0', 'undo_n', 'undo_n 3', 'unknown'])
      expect(parseSimpleSimonCommand(input, state)).toHaveProperty('error');
  });
});
