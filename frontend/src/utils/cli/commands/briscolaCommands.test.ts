import { describe, expect, it } from 'vitest';
import type { BriscolaResponse } from '../../../types/card';
import { parseBriscolaCommand } from './briscolaCommands';

const stateWithHandSize = (handSize: number): BriscolaResponse => ({
  players: [
    {
      id: 0,
      isHuman: true,
      cards: Array.from({ length: handSize }, () => ({ design: 'SPADE', value: 1 })),
      cardCount: handSize,
      points: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, cards: [], cardCount: 0, points: 0, trickCount: 0 },
  ],
  phase: 0,
  trickNumber: 1,
  currentPlayerIdx: 0,
  currentTrick: [],
  trumpSuit: 1,
  dealerIdx: 0,
  leadPlayerIdx: 0,
  stockRemaining: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: {},
});

describe('parseBriscolaCommand', () => {
  it('maps play commands to the same exec arguments as the card buttons', () => {
    expect(parseBriscolaCommand('p 0')).toEqual({ args: ['play', 0] });
    expect(parseBriscolaCommand('play 2')).toEqual({ args: ['play', 2] });
  });
  it('rejects missing, invalid, and out of range card indices', () => {
    for (const command of ['p', 'play x', 'p -1', 'p 3', 'p 1.5']) {
      expect(parseBriscolaCommand(command)).toEqual({ error: 'Usage: p <idx> (0-2)' });
    }
  });
  it('rejects p 2 when the human has two cards', () => {
    expect(parseBriscolaCommand('p 2', stateWithHandSize(2))).toEqual({ error: 'Usage: p <idx> (0-1)' });
  });
  it('maps each next alias to the page action', () => {
    expect(parseBriscolaCommand('n')).toEqual({ args: ['next'] });
    expect(parseBriscolaCommand('next')).toEqual({ args: ['next'] });
  });
  it('maps each hint alias to the server hint action', () => {
    expect(parseBriscolaCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBriscolaCommand('hint')).toEqual({ args: ['hint'] });
  });
  it('maps reset aliases to reset without config', () => {
    expect(parseBriscolaCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBriscolaCommand('reset')).toEqual({ args: ['reset'] });
  });
  it('rejects log and unknown commands', () => {
    expect(parseBriscolaCommand('log')).toEqual({ error: 'Unknown command: log' });
    expect(parseBriscolaCommand('deal')).toEqual({ error: 'Unknown command: deal' });
  });
});
