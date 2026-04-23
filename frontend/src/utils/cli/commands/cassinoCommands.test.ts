import { describe, expect, it } from 'vitest';
import type { Card, CassinoResponse } from '../../../types/card';
import { CASSINO_HELP, formatCassinoState, parseCassinoCommand } from './cassinoCommands';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<CassinoResponse> = {}): CassinoResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [],
        capturedCount: 0,
        sweepCount: 0,
        totalScore: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 2,
        cards: [],
        capturedCount: 1,
        sweepCount: 0,
        totalScore: 2,
      },
    ],
    currentTurn: 0,
    tableCards: [],
    builds: [],
    lastCaptureIdx: -1,
    gameEndFlag: false,
    phase: 'playerTurn',
    config: { targetScore: 21, multiBuildEnabled: true, sweepBonusEnabled: true, cpuDifficulty: 1 },
    cpuActions: [],
    humanAction: null,
    remainingDeck: 0,
    packsDealt: 1,
    roundWinners: [],
    lastRoundDetail: null,
    message: '',
    ...overrides,
  };
}

describe('parseCassinoCommand', () => {
  it('parses reset and r', () => {
    expect(parseCassinoCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseCassinoCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses next', () => {
    expect(parseCassinoCommand('n')).toEqual({ args: ['next'] });
  });

  it('parses log', () => {
    expect(parseCassinoCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses take with table indices', () => {
    expect(parseCassinoCommand('t 0 1 2')).toEqual({
      args: ['take', { handIndex: 0, tableIndices: [1, 2], buildIndices: [] }],
    });
  });

  it('parses take with build capture', () => {
    expect(parseCassinoCommand('t 0 b 1')).toEqual({
      args: ['take', { handIndex: 0, tableIndices: [], buildIndices: [1] }],
    });
  });

  it('errors for take without hand', () => {
    expect(parseCassinoCommand('t')).toHaveProperty('error');
  });

  it('errors for take with invalid hand index', () => {
    expect(parseCassinoCommand('t foo 1')).toHaveProperty('error');
  });

  it('parses build command', () => {
    expect(parseCassinoCommand('b 0 8 1')).toEqual({
      args: ['build', { handIndex: 0, tableIndices: [1], declaredValue: 8 }],
    });
  });

  it('errors on short build', () => {
    expect(parseCassinoCommand('b 0 8')).toHaveProperty('error');
  });

  it('errors on bad build value', () => {
    expect(parseCassinoCommand('b 0 foo 1')).toHaveProperty('error');
  });

  it('parses trail', () => {
    expect(parseCassinoCommand('tr 2')).toEqual({ args: ['trail', { handIndex: 2 }] });
  });

  it('errors on trail missing arg', () => {
    expect(parseCassinoCommand('tr')).toHaveProperty('error');
  });

  it('rejects unknown commands', () => {
    expect(parseCassinoCommand('foo')).toEqual({ error: 'Unknown command: foo' });
  });
});

describe('formatCassinoState', () => {
  it('renders turn and players', () => {
    const out = formatCassinoState(makeState());
    expect(out).toContain('Turn: Player 0');
    expect(out).toContain('You:');
    expect(out).toContain('CPU1:');
  });

  it('renders table cards line', () => {
    const out = formatCassinoState(makeState({ tableCards: [card('SPADE', 5), card('HEART', 8)] }));
    expect(out).toContain('Table:');
    expect(out).toContain('5SPADE');
    expect(out).toContain('8HEART');
  });

  it('renders builds', () => {
    const state = makeState({
      builds: [
        {
          ownerIdx: 0,
          value: 8,
          groups: [[card('SPADE', 3), card('HEART', 5)]],
          isMulti: false,
        },
      ],
    });
    const out = formatCassinoState(state);
    expect(out).toContain('Build');
    expect(out).toContain('val=8');
  });

  it('appends the message when present', () => {
    const out = formatCassinoState(makeState({ message: 'hello' }));
    expect(out).toContain('hello');
  });

  it('shows End when game ended', () => {
    const out = formatCassinoState(makeState({ gameEndFlag: true }));
    expect(out).toContain('Turn: End');
  });
});

describe('CASSINO_HELP', () => {
  it('mentions core commands', () => {
    const joined = CASSINO_HELP.join('\n');
    expect(joined).toContain('Take');
    expect(joined).toContain('build');
    expect(joined).toContain('Trail');
    expect(joined).toContain('reset');
  });
});
