import { describe, expect, it } from 'vitest';
import type { MonteCarloResponse } from '../../../types/card';
import { parseMonteCarloCommand } from './montecarloCommands';

const playing: MonteCarloResponse = {
  board: Array.from({ length: 5 }, (_, row) =>
    Array.from({ length: 5 }, (_, col) => ({
      card: row === 0 && col < 2 ? { design: col === 0 ? ('SPADE' as const) : ('HEART' as const), value: 7 } : null,
    })),
  ),
  phase: 0,
  stockCount: 1,
  removedCount: 0,
  dealCount: 0,
  removablePairs: 1,
  canUndo: true,
  isStalemate: false,
  message: '',
  messageCode: 'montecarlo.playing',
};

function filledState(): MonteCarloResponse {
  return {
    ...playing,
    board: playing.board.map((row) =>
      row.map((cell) => (cell.card ? cell : { card: { design: 'SPADE' as const, value: 7 } })),
    ),
  };
}

describe('parseMonteCarloCommand', () => {
  it('maps each remove syntax to the page button exec coordinates', () => {
    const rowOne = playing.board.map((row) => [...row]);
    const secondRow = rowOne[1];
    if (secondRow) secondRow[2] = { card: { design: 'SPADE', value: 7 } };
    expect(parseMonteCarloCommand('p 0 1', playing)).toEqual({ args: ['remove', 0, 0, 0, 1] });
    expect(parseMonteCarloCommand('play 1 7', { ...playing, board: rowOne })).toEqual({ args: ['remove', 0, 1, 1, 2] });
    const nearEnd = filledState();
    expect(parseMonteCarloCommand('remove 23 24', nearEnd)).toEqual({ args: ['remove', 4, 3, 4, 4] });
  });

  it('maps every other API command to the same values as its page button', () => {
    expect(parseMonteCarloCommand('d', playing)).toEqual({ args: ['deal'] });
    expect(parseMonteCarloCommand('deal', playing)).toEqual({ args: ['deal'] });
    expect(parseMonteCarloCommand('u', playing)).toEqual({ args: ['undo'] });
    expect(parseMonteCarloCommand('undo', playing)).toEqual({ args: ['undo'] });
    expect(parseMonteCarloCommand('g', playing)).toEqual({ args: ['giveup'] });
    expect(parseMonteCarloCommand('giveup', playing)).toEqual({ args: ['giveup'] });
    expect(parseMonteCarloCommand('h', playing)).toEqual({ args: ['hint'] });
    expect(parseMonteCarloCommand('hint', playing)).toEqual({ args: ['hint'] });
    expect(parseMonteCarloCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMonteCarloCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('rejects missing, invalid, out-of-range and extra cell indices', () => {
    for (const input of ['p', 'p 1', 'p x 2', 'p 1 x', 'p -1 2', 'p 25 1', 'p 4 4', 'p 0 1 2']) {
      expect(parseMonteCarloCommand(input, playing)).toHaveProperty('error');
    }
  });

  it('rejects illegal moves and unavailable actions with reasons', () => {
    const occupied = playing.board.map((row) => [...row]);
    const firstRow = occupied[0];
    const secondRow = occupied[1];
    if (firstRow) firstRow[2] = { card: { design: 'SPADE', value: 9 } };
    if (secondRow) secondRow[0] = { card: { design: 'SPADE', value: 9 } };
    expect(parseMonteCarloCommand('p 0 2', { ...playing, board: occupied })).toMatchObject({
      error: expect.stringContaining('adjacent'),
    });
    expect(parseMonteCarloCommand('p 0 5', { ...playing, board: occupied })).toMatchObject({
      error: expect.stringContaining('same rank'),
    });
    expect(parseMonteCarloCommand('p 2 3', playing)).toMatchObject({
      error: expect.stringContaining('contain a card'),
    });
    expect(parseMonteCarloCommand('d', { ...playing, stockCount: 0 })).toMatchObject({
      error: expect.stringContaining('stock'),
    });
    expect(parseMonteCarloCommand('u', { ...playing, canUndo: false })).toMatchObject({
      error: expect.stringContaining('undo'),
    });
    expect(parseMonteCarloCommand('p 0 1', { ...playing, phase: 2 })).toMatchObject({
      error: expect.stringContaining('in progress'),
    });
    expect(parseMonteCarloCommand('d', { ...playing, phase: 2 })).toHaveProperty('error');
    expect(parseMonteCarloCommand('u', { ...playing, phase: 2 })).toHaveProperty('error');
    expect(parseMonteCarloCommand('g', { ...playing, phase: 2 })).toHaveProperty('error');
    expect(parseMonteCarloCommand('h', { ...playing, phase: 2 })).toHaveProperty('error');
  });

  it('rejects unknown commands and does not expose action log', () => {
    expect(parseMonteCarloCommand('log', playing)).toHaveProperty('error');
    expect(parseMonteCarloCommand('xyz', playing)).toHaveProperty('error');
  });
});
