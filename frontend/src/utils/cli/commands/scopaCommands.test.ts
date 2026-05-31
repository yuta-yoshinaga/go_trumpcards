import { describe, expect, it } from 'vitest';
import type { ScopaResponse } from '../../../types/card';
import { formatScopaState, parseScopaCommand, SCOPA_HELP } from './scopaCommands';

describe('parseScopaCommand', () => {
  it('parses reset aliases to the short "r" command', () => {
    expect(parseScopaCommand('r')).toEqual({ args: ['r'] });
    expect(parseScopaCommand('reset')).toEqual({ args: ['r'] });
  });

  it('parses next aliases to the short "n" command', () => {
    expect(parseScopaCommand('n')).toEqual({ args: ['n'] });
    expect(parseScopaCommand('next')).toEqual({ args: ['n'] });
  });

  it('parses log aliases', () => {
    expect(parseScopaCommand('l')).toEqual({ args: ['log'] });
    expect(parseScopaCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses a play with table indices', () => {
    expect(parseScopaCommand('p 0 1 2')).toEqual({ args: ['p', { handIndex: 0, tableIndices: [1, 2] }] });
  });

  it('parses a play with no table indices as a lay', () => {
    expect(parseScopaCommand('p 0')).toEqual({ args: ['p', { handIndex: 0, tableIndices: [] }] });
  });

  it('accepts the "play" alias', () => {
    expect(parseScopaCommand('play 2 3')).toEqual({ args: ['p', { handIndex: 2, tableIndices: [3] }] });
  });

  it('errors when play has no hand index', () => {
    expect(parseScopaCommand('p')).toEqual({ error: 'Usage: p <hand> [tbl...]' });
  });

  it('errors on an invalid hand index', () => {
    expect(parseScopaCommand('p x')).toEqual({ error: 'Invalid hand index' });
  });

  it('errors on an invalid table index', () => {
    expect(parseScopaCommand('p 0 x')).toEqual({ error: 'Invalid table index' });
  });

  it('errors on an unknown command', () => {
    expect(parseScopaCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('errors on empty input', () => {
    expect(parseScopaCommand('')).toEqual({ error: 'Unknown command: ' });
  });
});

describe('formatScopaState', () => {
  const base: ScopaResponse = {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [],
        capturedCount: 2,
        scopaCount: 1,
        totalScore: 4,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], capturedCount: 0, scopaCount: 0, totalScore: 0 },
    ],
    currentTurn: 0,
    tableCards: [{ design: 'DIAMOND', value: 7 } as never],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 'playerTurn',
    config: { targetScore: 11, cpuDifficulty: 1 },
    cpuActions: [],
    humanAction: null,
    remainingDeck: 30,
    packsDealt: 1,
    roundWinners: [],
    lastRoundDetail: null,
    message: 'hi',
  } as ScopaResponse;

  it('renders players, table, and message', () => {
    const out = formatScopaState(base);
    expect(out).toContain('You: hand=3 captured=2 scopas=1 score=4');
    expect(out).toContain('CPU1: hand=3');
    expect(out).toContain('Table: 7DIAMOND');
    expect(out).toContain('hi');
  });

  it('marks the end turn when the game is over and omits empty table', () => {
    const out = formatScopaState({ ...base, gameEndFlag: true, tableCards: [], message: '' });
    expect(out).toContain('Turn: End');
    expect(out).not.toContain('Table:');
  });
});

describe('SCOPA_HELP', () => {
  it('lists the play command', () => {
    expect(SCOPA_HELP.some((l) => l.startsWith('p <hand>'))).toBe(true);
  });
});
