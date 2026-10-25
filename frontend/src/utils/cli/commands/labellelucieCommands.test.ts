import { describe, expect, it } from 'vitest';
import type { LaBelleLucieResponse } from '../../../types/card';
import { LaBelleLuciePhase } from '../../../types/phases';
import { parseLaBelleLucieCommand, toLaBelleLucieCommandState } from './labellelucieCommands';

const playingState = { fanCount: 18, phase: LaBelleLuciePhase.PLAYING, canUndo: true };

describe('parseLaBelleLucieCommand', () => {
  it('maps every supported command and alias to the page API operations', () => {
    const cases: [string, unknown[]][] = [
      ['p 2 4', ['mf', 2, 4]],
      ['p 0 17', ['mf', 0, 17]],
      ['p 17 4', ['mf', 17, 4]],
      ['play 2 4', ['mf', 2, 4]],
      ['f 3', ['ff', 3]],
      ['f 17', ['ff', 17]],
      ['foundation 3', ['ff', 3]],
      ['rd', ['rd']],
      ['redeal', ['rd']],
      ['ac', ['ac']],
      ['autocomplete', ['ac']],
      ['u', ['u']],
      ['undo', ['u']],
      ['undo_n 2', ['undo_n', undefined, undefined, 2]],
      ['g', ['giveup']],
      ['giveup', ['giveup']],
      ['h', ['hint']],
      ['hint', ['hint']],
      ['r', ['reset']],
      ['reset', ['reset']],
    ];
    for (const [input, args] of cases) expect(parseLaBelleLucieCommand(input, playingState)).toEqual({ args });
  });

  it('rejects missing arguments, invalid fan indices, and invalid undo counts', () => {
    for (const input of ['p', 'p 2', 'p -1 2', 'p 2 -1', 'p 18 2', 'p x 2'])
      expect(parseLaBelleLucieCommand(input)).toHaveProperty('error');
    for (const input of ['f', 'f -1', 'f 18', 'f x']) expect(parseLaBelleLucieCommand(input)).toHaveProperty('error');
    for (const input of ['undo_n', 'undo_n 0', 'undo_n -1', 'undo_n x'])
      expect(parseLaBelleLucieCommand(input)).toHaveProperty('error');
  });

  it('checks fan indices against the current fan count, which shrinks after a redeal', () => {
    expect(parseLaBelleLucieCommand('p 0 5', { ...playingState, fanCount: 6 })).toEqual({ args: ['mf', 0, 5] });
    expect(parseLaBelleLucieCommand('p 0 6', { ...playingState, fanCount: 6 })).toHaveProperty('error');
    expect(parseLaBelleLucieCommand('p 0 18', playingState)).toHaveProperty('error');
    expect(parseLaBelleLucieCommand('f 6', { ...playingState, fanCount: 6 })).toHaveProperty('error');
    expect(parseLaBelleLucieCommand('f 0', { ...playingState, fanCount: 0 })).toHaveProperty('error');
  });

  it('gates undo commands on an available undo in the playing phase', () => {
    for (const input of ['u', 'undo_n 2']) {
      expect(parseLaBelleLucieCommand(input, { ...playingState, canUndo: false })).toHaveProperty('error');
      expect(parseLaBelleLucieCommand(input, { ...playingState, phase: LaBelleLuciePhase.GAME_CLEAR })).toHaveProperty(
        'error',
      );
      expect(parseLaBelleLucieCommand(input)).toHaveProperty('error');
    }
    expect(parseLaBelleLucieCommand('u', playingState)).toEqual({ args: ['u'] });
    expect(parseLaBelleLucieCommand('undo_n 2', playingState)).toEqual({ args: ['undo_n', undefined, undefined, 2] });
  });

  it('rejects unknown commands without accepting log', () => {
    expect(parseLaBelleLucieCommand('bogus')).toEqual({ error: 'Unknown command: bogus' });
    expect(parseLaBelleLucieCommand('log')).toHaveProperty('error');
  });

  it('rejects fan moves before the state has loaded', () => {
    expect(parseLaBelleLucieCommand('p 0 1')).toHaveProperty('error');
    expect(parseLaBelleLucieCommand('f 0')).toHaveProperty('error');
  });
});

describe('toLaBelleLucieCommandState', () => {
  it('refuses everything state-dependent before the state loads', () => {
    expect(toLaBelleLucieCommandState(null)).toEqual({
      fanCount: 0,
      phase: LaBelleLuciePhase.GAME_OVER,
      canUndo: false,
    });
  });

  it('maps the fan count, phase and undo availability from the page state', () => {
    const state = {
      fans: [[], [], []],
      phase: LaBelleLuciePhase.PLAYING,
      canUndo: true,
    } as unknown as LaBelleLucieResponse;
    expect(toLaBelleLucieCommandState(state)).toEqual({ fanCount: 3, phase: LaBelleLuciePhase.PLAYING, canUndo: true });
  });
});
