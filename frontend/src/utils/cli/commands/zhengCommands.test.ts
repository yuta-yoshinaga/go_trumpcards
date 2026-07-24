import { describe, expect, it } from 'vitest';
import type { ZhengResponse } from '../../../types/card';
import { formatZhengState, parseZhengCommand, ZHENG_HELP } from './zhengCommands';

describe('parseZhengCommand', () => {
  it('parses play with indices', () => {
    expect(parseZhengCommand('p 0 2')).toEqual({ args: ['play', [0, 2]] });
    expect(parseZhengCommand('play 1')).toEqual({ args: ['play', [1]] });
  });

  it('treats play with no args as a pass', () => {
    expect(parseZhengCommand('p')).toEqual({ args: ['play'] });
  });

  it('reports invalid indices', () => {
    const res = parseZhengCommand('p abc');
    expect('error' in res && res.error).toContain('Invalid indices');
  });

  it('parses reset', () => {
    expect(parseZhengCommand('r')).toEqual({ args: ['reset'] });
    expect(parseZhengCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command', () => {
    const res = parseZhengCommand('rese');
    expect('error' in res && res.error).toContain('Did you mean');
  });

  it('reports an unknown command with no suggestion', () => {
    const res = parseZhengCommand('zzz');
    expect('error' in res && res.error).toContain('Unknown command');
  });
});

describe('formatZhengState', () => {
  const base: ZhengResponse = {
    players: [
      { id: 0, isHuman: true, isFinished: false, rank: 0, cardCount: 14, cards: [] },
      { id: 1, isHuman: false, isFinished: true, rank: 1, cardCount: 0, cards: [] },
    ],
    currentTurn: 0,
    tableCards: [],
    tablePlayType: 0,
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    cpuActions: [],
    humanAction: null,
    config: { cpuDifficulty: 0 },
    message: '',
  };

  it('formats players and turn', () => {
    const out = formatZhengState(base);
    expect(out).toContain('Turn: Player 0');
    expect(out).toContain('You: 14 cards');
    expect(out).toContain('CPU1: rank=1');
  });

  it('formats table cards, end turn, and message', () => {
    const out = formatZhengState({
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
    expect(ZHENG_HELP.length).toBeGreaterThan(0);
  });
});
