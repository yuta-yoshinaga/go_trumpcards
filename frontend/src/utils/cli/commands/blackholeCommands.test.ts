import { describe, expect, it } from 'vitest';
import type { Card } from '../../../types/card';
import type { BlackHoleResponse } from '../../../types/games/blackhole';
import { BlackHolePhase } from '../../../types/phases';
import type { CliParseResult } from '../types';
import { BLACKHOLE_HELP, type BlackHoleCliArgs, getLegalFans, parseBlackHoleCommand } from './blackholeCommands';

/** The error text of a parse result, or undefined when it parsed successfully. */
function errorOf(res: CliParseResult<BlackHoleCliArgs>): string | undefined {
  return 'error' in res ? res.error : undefined;
}

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeMockState(overrides: Partial<BlackHoleResponse> = {}): BlackHoleResponse {
  const fans: Card[][] = Array.from({ length: 17 }, () => []);
  // fan 0: top is 2 (playable onto ♠A which is 1)
  fans[0] = [card('HEART', 9), card('CLOVER', 2)];
  // fan 1: top is 5 (not playable onto ♠A)
  fans[1] = [card('DIAMOND', 5)];
  // fan 2: empty
  fans[2] = [];
  return {
    fans,
    blackHole: [card('SPADE', 1)],
    phase: BlackHolePhase.PLAYING,
    moveCount: 0,
    canUndo: true,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

describe('getLegalFans', () => {
  it('identifies fans whose top card rank is ±1 from the black hole top', () => {
    const state = makeMockState();
    const legal = getLegalFans(state);
    expect(legal.has(0)).toBe(true);
    expect(legal.has(1)).toBe(false);
    expect(legal.has(2)).toBe(false);
  });

  it('returns empty set when black hole has no cards', () => {
    const legal = getLegalFans({ fans: [[card('SPADE', 2)]], blackHole: [] });
    expect(legal.size).toBe(0);
  });

  it('handles empty fans or unpopulated fans', () => {
    const legal = getLegalFans({ fans: [[], []], blackHole: [card('SPADE', 1)] });
    expect(legal.size).toBe(0);
  });
});

describe('parseBlackHoleCommand without state', () => {
  it('parses play command aliases to the exact button arguments for exec("mb", { fan })', () => {
    for (const cmd of ['p 3', 'play 3', 'mb 3', 'm 3', 'move 3']) {
      expect(parseBlackHoleCommand(cmd)).toEqual({ args: ['mb', { fan: 3 }] });
    }
  });

  it('rejects play without numeric argument', () => {
    expect(parseBlackHoleCommand('p')).toEqual({ error: 'Usage: p <fan>' });
    expect(parseBlackHoleCommand('play abc')).toEqual({ error: 'Usage: p <fan>' });
  });

  it('rejects negative fan index when state is omitted', () => {
    expect(parseBlackHoleCommand('p -1')).toEqual({ error: 'Invalid fan index' });
  });

  it('parses undo command aliases to exec("u")', () => {
    expect(parseBlackHoleCommand('u')).toEqual({ args: ['u'] });
    expect(parseBlackHoleCommand('undo')).toEqual({ args: ['u'] });
  });

  it('parses giveup command aliases to exec("g")', () => {
    expect(parseBlackHoleCommand('g')).toEqual({ args: ['g'] });
    expect(parseBlackHoleCommand('giveup')).toEqual({ args: ['g'] });
  });

  it('parses hint command aliases to exec("hint")', () => {
    expect(parseBlackHoleCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBlackHoleCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset command aliases to exec("reset")', () => {
    expect(parseBlackHoleCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBlackHoleCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('reports an unknown command and suggests near miss', () => {
    expect(errorOf(parseBlackHoleCommand('pla'))).toBe('Unknown command: pla. Did you mean: play?');
    expect(parseBlackHoleCommand('xyz')).toEqual({ error: 'Unknown command: xyz' });
  });
});

describe('parseBlackHoleCommand with state validation', () => {
  it('accepts legal moves matching the page button exec("mb", { fan })', () => {
    const state = makeMockState();
    expect(parseBlackHoleCommand('p 0', state)).toEqual({ args: ['mb', { fan: 0 }] });
  });

  it('rejects moves when out of range or targeting an empty fan', () => {
    const state = makeMockState();
    expect(parseBlackHoleCommand('p -1', state)).toEqual({ error: 'Fan index out of range (0-16)' });
    expect(parseBlackHoleCommand('p 17', state)).toEqual({ error: 'Fan index out of range (0-16)' });
    expect(parseBlackHoleCommand('p 2', state)).toEqual({ error: 'Fan 2 is empty' });
  });

  it('rejects moves whose top card is not adjacent in rank', () => {
    const state = makeMockState();
    expect(parseBlackHoleCommand('p 1', state)).toEqual({
      error: 'Fan 1 cannot be moved to the black hole',
    });
  });

  it('rejects moves when game is not playing', () => {
    const clearState = makeMockState({ phase: BlackHolePhase.GAME_CLEAR });
    expect(parseBlackHoleCommand('p 0', clearState)).toEqual({ error: 'Game is not in playing phase' });
    const overState = makeMockState({ phase: BlackHolePhase.GAME_OVER });
    expect(parseBlackHoleCommand('p 0', overState)).toEqual({ error: 'Game is not in playing phase' });
  });

  it('validates undo against game phase and canUndo flag', () => {
    const state = makeMockState({ canUndo: true });
    expect(parseBlackHoleCommand('u', state)).toEqual({ args: ['u'] });

    const noUndoState = makeMockState({ canUndo: false });
    expect(parseBlackHoleCommand('u', noUndoState)).toEqual({ error: 'No moves to undo' });

    const overState = makeMockState({ phase: BlackHolePhase.GAME_OVER, canUndo: true });
    expect(parseBlackHoleCommand('u', overState)).toEqual({ error: 'Game is not in playing phase' });
  });

  it('validates giveup against game phase', () => {
    const state = makeMockState();
    expect(parseBlackHoleCommand('g', state)).toEqual({ args: ['g'] });

    const overState = makeMockState({ phase: BlackHolePhase.GAME_OVER });
    expect(parseBlackHoleCommand('g', overState)).toEqual({ error: 'Game is not in playing phase' });
  });

  it('validates hint against game phase', () => {
    const state = makeMockState();
    expect(parseBlackHoleCommand('h', state)).toEqual({ args: ['hint'] });

    const clearState = makeMockState({ phase: BlackHolePhase.GAME_CLEAR });
    expect(parseBlackHoleCommand('h', clearState)).toEqual({ error: 'Game is not in playing phase' });
  });

  it('always allows reset even if game is over', () => {
    const overState = makeMockState({ phase: BlackHolePhase.GAME_OVER });
    expect(parseBlackHoleCommand('r', overState)).toEqual({ args: ['reset'] });
  });

  it('verifies help text content', () => {
    const help = BLACKHOLE_HELP.join('\n');
    expect(help).toContain('mb');
    expect(help).toContain('undo');
    expect(help).toContain('hint');
    expect(help).toContain('giveup');
    expect(help).toContain('reset');
  });
});
