import { describe, expect, it } from 'vitest';
import { GLEEK_HELP, parseGleekCommand } from './gleekCommands';

describe('parseGleekCommand', () => {
  it('parses a raise', () => {
    expect(parseGleekCommand('bid 14')).toEqual({ args: ['bid', { bid: 14 }] });
    expect(parseGleekCommand('b 16')).toEqual({ args: ['bid', { bid: 16 }] });
  });

  it('parses dropping out', () => {
    expect(parseGleekCommand('bid pass')).toEqual({ args: ['bid', { bid: 0 }] });
    expect(parseGleekCommand('b p')).toEqual({ args: ['bid', { bid: 0 }] });
  });

  it('returns error for a bid without a usable amount', () => {
    expect('error' in parseGleekCommand('bid')).toBe(true);
    expect('error' in parseGleekCommand('bid twelve')).toBe(true);
    expect('error' in parseGleekCommand('bid -2')).toBe(true);
  });

  // **捨て札フェーズを抜ける唯一の入力。** これが無いと落札の直後に CLI から
  // 先へ進めなくなる。
  it('parses the discard list, space- or comma-separated', () => {
    expect(parseGleekCommand('discard 0 1 2 3 4 5 6')).toEqual({
      args: ['discard', { discardIndices: [0, 1, 2, 3, 4, 5, 6] }],
    });
    expect(parseGleekCommand('d 0,1,2')).toEqual({ args: ['discard', { discardIndices: [0, 1, 2] }] });
  });

  it('returns error for a discard with no or bad indices', () => {
    expect('error' in parseGleekCommand('discard')).toBe(true);
    expect('error' in parseGleekCommand('d 0 x 2')).toBe(true);
    expect('error' in parseGleekCommand('d -1')).toBe(true);
  });

  it('parses play (short and long)', () => {
    expect(parseGleekCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseGleekCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseGleekCommand('p')).toBe(true);
  });

  it('parses next, nextround, hint and reset', () => {
    expect(parseGleekCommand('n')).toEqual({ args: ['next'] });
    expect(parseGleekCommand('next')).toEqual({ args: ['next'] });
    expect(parseGleekCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseGleekCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseGleekCommand('h')).toEqual({ args: ['hint'] });
    expect(parseGleekCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseGleekCommand('r')).toEqual({ args: ['reset'] });
    expect(parseGleekCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseGleekCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseGleekCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text covering both input phases', () => {
    expect(GLEEK_HELP.some((line) => line.includes('Drop out'))).toBe(true);
    expect(GLEEK_HELP.some((line) => line.includes('discard'))).toBe(true);
    expect(GLEEK_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
