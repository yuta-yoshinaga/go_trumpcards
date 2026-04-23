import { describe, expect, it } from 'vitest';
import type { Card, PresidentResponse } from '../../../types/card';
import { formatPresidentState, PRESIDENT_HELP, parsePresidentCommand } from './presidentCommands';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<PresidentResponse> = {}): PresidentResponse {
  return {
    players: [
      { id: 0, isHuman: true, isFinished: false, rank: -1, cardCount: 3, cards: [] },
      { id: 1, isHuman: false, isFinished: false, rank: -1, cardCount: 5, cards: [] },
    ],
    currentTurn: 0,
    tableCards: [],
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    revolutionActive: false,
    config: {
      revolutionEnabled: true,
      cardExchangeEnabled: true,
      passFieldFlushEnabled: true,
      cpuDifficulty: 1,
    },
    exchangeActions: [],
    cpuActions: [],
    humanAction: null,
    message: '',
    ...overrides,
  };
}

describe('parsePresidentCommand', () => {
  it('parses reset and its alias r', () => {
    expect(parsePresidentCommand('reset')).toEqual({ args: ['reset'] });
    expect(parsePresidentCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses play with no indices as pass', () => {
    expect(parsePresidentCommand('p')).toEqual({ args: ['play', []] });
    expect(parsePresidentCommand('play')).toEqual({ args: ['play', []] });
  });

  it('parses play with indices', () => {
    expect(parsePresidentCommand('p 0 1 2')).toEqual({ args: ['play', [0, 1, 2]] });
    expect(parsePresidentCommand('play 3')).toEqual({ args: ['play', [3]] });
  });

  it('rejects play with non-numeric indices', () => {
    expect(parsePresidentCommand('p foo')).toEqual({ error: 'Usage: p [idx ...]' });
  });

  it('parses log and its alias l', () => {
    expect(parsePresidentCommand('log')).toEqual({ args: ['log'] });
    expect(parsePresidentCommand('l')).toEqual({ args: ['log'] });
  });

  it('rejects unknown commands', () => {
    const result = parsePresidentCommand('foo');
    expect(result).toEqual({ error: 'Unknown command: foo' });
  });

  it('treats empty input as unknown command', () => {
    const result = parsePresidentCommand('');
    expect(result).toEqual({ error: 'Unknown command: ' });
  });

  it('handles leading/trailing whitespace and upper-case', () => {
    expect(parsePresidentCommand('  PLAY  0  ')).toEqual({ args: ['play', [0]] });
  });
});

describe('formatPresidentState', () => {
  it('renders basic play state', () => {
    const out = formatPresidentState(makeState());
    expect(out).toContain('Turn: Player 0');
    expect(out).toContain('You: 3 cards');
    expect(out).toContain('CPU1: 5 cards');
    expect(out).not.toContain('REVOLUTION');
  });

  it('renders revolution banner', () => {
    const out = formatPresidentState(makeState({ revolutionActive: true }));
    expect(out).toContain('*** REVOLUTION ACTIVE ***');
  });

  it('renders game-end with ranks', () => {
    const out = formatPresidentState(
      makeState({
        gameEndFlag: true,
        players: [
          { id: 0, isHuman: true, isFinished: true, rank: 1, cardCount: 0, cards: [] },
          { id: 1, isHuman: false, isFinished: true, rank: 2, cardCount: 0, cards: [] },
        ],
      }),
    );
    expect(out).toContain('Turn: End');
    expect(out).toContain('You: rank=1');
    expect(out).toContain('CPU1: rank=2');
  });

  it('renders the table cards line when cards are present', () => {
    const out = formatPresidentState(
      makeState({
        tableCards: [card('SPADE', 5), card('HEART', 5)],
      }),
    );
    expect(out).toMatch(/Table: 5SPADE 5HEART/);
  });

  it('omits table line when empty', () => {
    const out = formatPresidentState(makeState());
    expect(out).not.toContain('Table:');
  });

  it('appends the message when present', () => {
    const out = formatPresidentState(makeState({ message: 'hello' }));
    expect(out).toContain('hello');
  });
});

describe('PRESIDENT_HELP', () => {
  it('exposes the expected commands', () => {
    const joined = PRESIDENT_HELP.join('\n');
    expect(joined).toContain('Play');
    expect(joined).toContain('reset');
    expect(joined).toContain('log');
  });
});
