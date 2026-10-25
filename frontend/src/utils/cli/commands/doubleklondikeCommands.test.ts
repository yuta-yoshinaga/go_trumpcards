import { describe, expect, it } from 'vitest';
import type { DoubleKlondikeResponse } from '../../../types/card';
import { DoubleKlondikePhase } from '../../../types/phases';
import { parseDoubleKlondikeCommand } from './doubleklondikeCommands';

const state: DoubleKlondikeResponse = {
  tableau: [
    [
      { card: { design: 'SPADE', value: 9 }, faceUp: true },
      { card: { design: 'HEART', value: 8 }, faceUp: true },
    ],
    [],
    [],
  ],
  stockCount: 2,
  waste: [{ design: 'HEART', value: 1 }],
  foundation: Array.from({ length: 8 }, () => []),
  phase: DoubleKlondikePhase.PLAYING,
  moveCount: 0,
  canUndo: true,
  isStalemate: false,
  legalTargets: { waste: { tableau: [1], foundation: [0] }, 'tableau-0-1': { tableau: [2], foundation: [0] } },
  message: '',
};

describe('parseDoubleKlondikeCommand', () => {
  it.each([
    ['p waste f', ['mwf']],
    ['p waste 1', ['mwt', { col: 1 }]],
    ['p 0 1 f', ['mtf', { col: 0 }]],
    ['p 0 1 2', ['mtt', { fromCol: 0, cardIndex: 1, toCol: 2 }]],
    ['d', ['d']],
    ['draw', ['d']],
    ['ac', ['ac']],
    ['auto', ['ac']],
    ['u', ['u']],
    ['undo', ['u']],
    ['h', ['hint']],
    ['hint', ['hint']],
    ['g', ['g']],
    ['giveup', ['g']],
    ['r', ['reset']],
    ['reset', ['reset']],
  ])('%s matches the page action', (input, args) => expect(parseDoubleKlondikeCommand(input, state)).toEqual({ args }));
  it.each(['p', 'p waste', 'p 0 1', 'p x f'])('rejects incomplete or malformed move %s', (input) =>
    expect(parseDoubleKlondikeCommand(input, state)).toHaveProperty('error'),
  );
  it.each(['p waste 3', 'p 0 8 2', 'p 4 0 f', 'p 0 1 9'])('rejects out of range or illegal move %s', (input) =>
    expect(parseDoubleKlondikeCommand(input, state)).toHaveProperty('error'),
  );
  it('rejects unknown commands', () => {
    expect(parseDoubleKlondikeCommand('wat', state)).toHaveProperty(
      'error',
      expect.stringContaining('Unknown command: wat'),
    );
  });
  it('allows draw without state, but rejects an exhausted draw', () => {
    expect(parseDoubleKlondikeCommand('d')).toEqual({ args: ['d'] });
    expect(parseDoubleKlondikeCommand('d', { ...state, stockCount: 0, waste: [] })).toHaveProperty('error');
  });
  it('requires undo to be available and gates actions after game end', () => {
    expect(parseDoubleKlondikeCommand('u', { ...state, canUndo: false })).toHaveProperty('error');
    for (const command of ['d', 'ac', 'u', 'h', 'g'])
      expect(parseDoubleKlondikeCommand(command, { ...state, phase: DoubleKlondikePhase.GAME_OVER })).toHaveProperty(
        'error',
      );
  });
});
