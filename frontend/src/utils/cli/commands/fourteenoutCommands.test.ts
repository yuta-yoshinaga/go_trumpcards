import { describe, expect, it } from 'vitest';
import type { FourteenOutResponse } from '../../../types/card';
import { FourteenOutPhase } from '../../../types/phases';
import { parseFourteenOutCommand } from './fourteenoutCommands';

const state: FourteenOutResponse = {
  columns: [
    [{ card: { design: 'SPADE', value: 9 } }, { card: { design: 'CLOVER', value: 2 } }],
    [{ card: { design: 'HEART', value: 5 } }],
    [],
    ...Array.from({ length: 9 }, () => []),
  ],
  phase: FourteenOutPhase.PLAYING,
  removedCount: 0,
  removablePairs: 1,
  legalTargets: { 'tableau-0': { tableau: [1], foundation: [] } },
  canUndo: true,
  isStalemate: false,
  message: '',
  messageCode: 'fourteenout.playing',
};

describe('parseFourteenOutCommand', () => {
  it('maps each supported action and aliases to page exec arguments', () => {
    expect(parseFourteenOutCommand('p 0 1', state)).toEqual({ args: ['remove', 0, 1] });
    expect(parseFourteenOutCommand('play 0 1', state)).toEqual({ args: ['remove', 0, 1] });
    expect(parseFourteenOutCommand('u', state)).toEqual({ args: ['undo'] });
    expect(parseFourteenOutCommand('undo', state)).toEqual({ args: ['undo'] });
    expect(parseFourteenOutCommand('g', state)).toEqual({ args: ['giveup'] });
    expect(parseFourteenOutCommand('giveup', state)).toEqual({ args: ['giveup'] });
    expect(parseFourteenOutCommand('h', state)).toEqual({ args: ['hint'] });
    expect(parseFourteenOutCommand('hint', state)).toEqual({ args: ['hint'] });
    expect(parseFourteenOutCommand('r', state)).toEqual({ args: ['reset'] });
    expect(parseFourteenOutCommand('reset', state)).toEqual({ args: ['reset'] });
  });

  it('rejects missing, malformed, out of range, empty, illegal and same columns', () => {
    for (const command of ['p', 'p 0', 'p x 1'])
      expect(parseFourteenOutCommand(command, state)).toHaveProperty('error');
    for (const command of ['p -1 1', 'p 12 1', 'p 0 12'])
      expect(parseFourteenOutCommand(command, state)).toHaveProperty('error');
    expect(parseFourteenOutCommand('p 0 0', state)).toHaveProperty('error');
    expect(parseFourteenOutCommand('p 0 2', state)).toHaveProperty('error');
    expect(parseFourteenOutCommand('p 0 3', state)).toHaveProperty('error');
  });

  it('rejects play and guarded actions outside their valid state', () => {
    expect(parseFourteenOutCommand('p 0 1')).toHaveProperty('error');
    expect(parseFourteenOutCommand('p 0 1', { ...state, phase: FourteenOutPhase.GAME_OVER })).toHaveProperty('error');
    expect(parseFourteenOutCommand('u', { ...state, canUndo: false })).toHaveProperty('error');
    expect(parseFourteenOutCommand('g', { ...state, phase: FourteenOutPhase.GAME_CLEAR })).toHaveProperty('error');
    expect(parseFourteenOutCommand('zz')).toHaveProperty('error');
    expect(parseFourteenOutCommand('log')).toHaveProperty('error');
  });
});
