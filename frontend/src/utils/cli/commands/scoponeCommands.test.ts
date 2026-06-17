import { describe, expect, it } from 'vitest';
import { makeScoponeState } from '../../../test/stateFactories';
import type { ScoponeResponse } from '../../../types/card';
import { formatScoponeState, parseScoponeCommand, SCOPONE_HELP } from './scoponeCommands';

describe('parseScoponeCommand', () => {
  it('parses reset aliases to the short "r" command', () => {
    expect(parseScoponeCommand('r')).toEqual({ args: ['r'] });
    expect(parseScoponeCommand('reset')).toEqual({ args: ['r'] });
  });

  it('parses next aliases to the short "n" command', () => {
    expect(parseScoponeCommand('n')).toEqual({ args: ['n'] });
    expect(parseScoponeCommand('next')).toEqual({ args: ['n'] });
  });

  it('parses log aliases', () => {
    expect(parseScoponeCommand('l')).toEqual({ args: ['log'] });
    expect(parseScoponeCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses a play with table indices', () => {
    expect(parseScoponeCommand('p 0 1 2')).toEqual({ args: ['p', { handIndex: 0, tableIndices: [1, 2] }] });
  });

  it('parses a play with no table indices as a lay', () => {
    expect(parseScoponeCommand('p 0')).toEqual({ args: ['p', { handIndex: 0, tableIndices: [] }] });
  });

  it('accepts the "play" alias', () => {
    expect(parseScoponeCommand('play 2 3')).toEqual({ args: ['p', { handIndex: 2, tableIndices: [3] }] });
  });

  it('errors when play has no hand index', () => {
    expect(parseScoponeCommand('p')).toEqual({ error: 'Usage: p <hand> [tbl...]' });
  });

  it('errors on an invalid hand index', () => {
    expect(parseScoponeCommand('p x')).toEqual({ error: 'Invalid hand index' });
  });

  it('errors on an invalid table index', () => {
    expect(parseScoponeCommand('p 0 x')).toEqual({ error: 'Invalid table index' });
  });

  it('errors on an unknown command', () => {
    expect(parseScoponeCommand('zzz')).toEqual({ error: 'Unknown command: zzz' });
  });

  it('errors on empty input', () => {
    expect(parseScoponeCommand('')).toEqual({ error: 'Unknown command: ' });
  });
});

describe('formatScoponeState', () => {
  const base: ScoponeResponse = makeScoponeState({
    players: [
      { id: 0, isHuman: true, team: 0, handCount: 3, cards: [], capturedCount: 2, scopaCount: 1 },
      { id: 1, isHuman: false, team: 1, handCount: 3, cards: [], capturedCount: 0, scopaCount: 0 },
      { id: 2, isHuman: false, team: 0, handCount: 3, cards: [], capturedCount: 0, scopaCount: 0 },
      { id: 3, isHuman: false, team: 1, handCount: 3, cards: [], capturedCount: 0, scopaCount: 0 },
    ],
    tableCards: [{ design: 'DIAMOND', value: 7 }],
    teamScores: [4, 2],
    message: 'hi',
  });

  it('renders players, teams, table, scores, and message', () => {
    const out = formatScoponeState(base);
    expect(out).toContain('P0 (You) team0: hand=3 captured=2 scopas=1');
    expect(out).toContain('P1 team1: hand=3');
    expect(out).toContain('Table: 7DIAMOND');
    expect(out).toContain('Team0: 4');
    expect(out).toContain('Team1: 2');
    expect(out).toContain('hi');
  });

  it('marks the end turn when the game is over and shows empty table', () => {
    const out = formatScoponeState(makeScoponeState({ gameEndFlag: true, tableCards: [], message: '' }));
    expect(out).toContain('Turn: End');
    expect(out).toContain('Table: (empty)');
  });
});

describe('SCOPONE_HELP', () => {
  it('lists the play command', () => {
    expect(SCOPONE_HELP.some((l) => l.startsWith('p <hand>'))).toBe(true);
  });
});
