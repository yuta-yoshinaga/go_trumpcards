import { describe, expect, it } from 'vitest';
import type { TienLenResponse } from '../../../types/card';
import { formatTienLenState, parseTienLenCommand, TIENLEN_HELP } from './tienlenCommands';

describe('parseTienLenCommand', () => {
  it('parses play with indices', () => {
    expect(parseTienLenCommand('p 0 2')).toEqual({ args: ['play', [0, 2]] });
    expect(parseTienLenCommand('play 1')).toEqual({ args: ['play', [1]] });
  });

  it('treats play with no args as a pass', () => {
    expect(parseTienLenCommand('p')).toEqual({ args: ['play'] });
  });

  it('reports invalid indices', () => {
    const res = parseTienLenCommand('p abc');
    expect('error' in res && res.error).toContain('Invalid indices');
  });

  it('parses reset', () => {
    expect(parseTienLenCommand('r')).toEqual({ args: ['reset'] });
    expect(parseTienLenCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command', () => {
    const res = parseTienLenCommand('rese');
    expect('error' in res && res.error).toContain('Did you mean');
  });

  it('reports an unknown command with no suggestion', () => {
    const res = parseTienLenCommand('zzz');
    expect('error' in res && res.error).toContain('Unknown command');
  });
});

describe('formatTienLenState', () => {
  const base: TienLenResponse = {
    players: [
      { id: 0, isHuman: true, isFinished: false, rank: 0, cardCount: 13, cards: [] },
      { id: 1, isHuman: false, isFinished: true, rank: 1, cardCount: 0, cards: [] },
    ],
    currentTurn: 0,
    tableCards: [],
    tablePlayType: 0,
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    cpuActions: [],
    humanAction: null,
    config: { cpuDifficulty: 1 },
    message: '',
  };

  it('formats players and turn', () => {
    const out = formatTienLenState(base);
    expect(out).toContain('Turn: Player 0');
    expect(out).toContain('You: 13 cards');
    expect(out).toContain('CPU1: rank=1');
  });

  it('formats table cards, end turn, and message', () => {
    const out = formatTienLenState({
      ...base,
      gameEndFlag: true,
      tableCards: [{ design: 'SPADE', value: 3 }],
      message: 'done',
    });
    expect(out).toContain('Turn: End');
    expect(out).toContain('Table: 3SPADE');
    expect(out).toContain('done');
  });

  it('exposes help text', () => {
    expect(TIENLEN_HELP.length).toBeGreaterThan(0);
  });
});
