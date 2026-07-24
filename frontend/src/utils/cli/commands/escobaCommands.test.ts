import { describe, expect, it } from 'vitest';
import { makeEscobaState } from '../../../test/stateFactories';
import type { EscobaResponse } from '../../../types/card';
import { ESCOBA_HELP, formatEscobaState, parseEscobaCommand } from './escobaCommands';

describe('parseEscobaCommand', () => {
  it('parses reset aliases to the short "r" command', () => {
    expect(parseEscobaCommand('r')).toEqual({ args: ['r'] });
    expect(parseEscobaCommand('reset')).toEqual({ args: ['r'] });
  });

  it('parses next aliases to the short "n" command', () => {
    expect(parseEscobaCommand('n')).toEqual({ args: ['n'] });
    expect(parseEscobaCommand('next')).toEqual({ args: ['n'] });
  });

  it('parses log aliases', () => {
    expect(parseEscobaCommand('l')).toEqual({ args: ['log'] });
    expect(parseEscobaCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses a play with table indices', () => {
    expect(parseEscobaCommand('p 0 1 2')).toEqual({ args: ['p', { handIndex: 0, tableIndices: [1, 2] }] });
  });

  it('parses a play with no table indices as a lay', () => {
    expect(parseEscobaCommand('p 0')).toEqual({ args: ['p', { handIndex: 0, tableIndices: [] }] });
  });

  it('accepts the "play" alias', () => {
    expect(parseEscobaCommand('play 2 3')).toEqual({ args: ['p', { handIndex: 2, tableIndices: [3] }] });
  });

  it('errors when play has no hand index', () => {
    expect(parseEscobaCommand('p')).toEqual({ error: 'Usage: p <hand> [tbl...]' });
  });

  it('errors on an invalid hand index', () => {
    expect(parseEscobaCommand('p x')).toEqual({ error: 'Invalid hand index' });
  });

  it('errors on an invalid table index', () => {
    expect(parseEscobaCommand('p 0 x')).toEqual({ error: 'Invalid table index' });
  });

  it('errors on an unknown command', () => {
    expect(parseEscobaCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('errors on empty input', () => {
    expect(parseEscobaCommand('')).toEqual({ error: 'Unknown command: ' });
  });
});

describe('formatEscobaState', () => {
  const base: EscobaResponse = makeEscobaState({
    players: [
      { id: 0, isHuman: true, handCount: 3, cards: [], capturedCount: 2, capturedCards: [], escobaCount: 1, score: 4 },
      { id: 1, isHuman: false, handCount: 3, cards: [], capturedCount: 0, capturedCards: [], escobaCount: 0, score: 2 },
      { id: 2, isHuman: false, handCount: 3, cards: [], capturedCount: 0, capturedCards: [], escobaCount: 0, score: 0 },
      { id: 3, isHuman: false, handCount: 3, cards: [], capturedCount: 0, capturedCards: [], escobaCount: 0, score: 0 },
    ],
    tableCards: [{ design: 'SPADE', value: 7 }],
    stockRemaining: 16,
    message: 'hi',
  });

  it('renders players, table, stock, and message', () => {
    const out = formatEscobaState(base);
    expect(out).toContain('P0 (You): hand=3 captured=2 escobas=1 score=4');
    expect(out).toContain('P1: hand=3');
    expect(out).toContain('Table: 7SPADE');
    expect(out).toContain('Stock: 16');
    expect(out).toContain('hi');
  });

  it('marks the end turn when the game is over and shows empty table', () => {
    const out = formatEscobaState(makeEscobaState({ gameEndFlag: true, tableCards: [], message: '' }));
    expect(out).toContain('Turn: End');
    expect(out).toContain('Table: (empty)');
  });
});

describe('ESCOBA_HELP', () => {
  it('lists the play command', () => {
    expect(ESCOBA_HELP.some((l) => l.startsWith('p <hand>'))).toBe(true);
  });
});
