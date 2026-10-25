import { describe, expect, it } from 'vitest';
import type { SimpleSimonResponse } from '../../../types/card';
import { parseSimpleSimonCommand } from './simplesimonCommands';

const state = {
  columns: [[{ design: 'SPADE', value: 3 }], []],
  runStarts: [[0], []],
  legalTargets: { 'tableau-0-0': { tableau: [1], foundation: [] } },
  phase: 0,
  moveCount: 2,
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
  it('rejects incomplete, out of range and illegal inputs', () => {
    for (const input of ['m', 'm 0 0', 'm -1 0 1', 'm 0 4 1', 'm 0 0 9', 'm 1 0 0', 'undo_n', 'undo_n 3', 'unknown'])
      expect(parseSimpleSimonCommand(input, state)).toHaveProperty('error');
  });
});
