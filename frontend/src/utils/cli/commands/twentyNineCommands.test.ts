import { describe, expect, it } from 'vitest';
import { parseTwentyNineCommand, TWENTY_NINE_HELP } from './twentyNineCommands';

describe('parseTwentyNineCommand', () => {
  it('parses bid (short and long)', () => {
    expect(parseTwentyNineCommand('b 16')).toEqual({ args: ['bid', { bid: 16 }] });
    expect(parseTwentyNineCommand('bid 28')).toEqual({ args: ['bid', { bid: 28 }] });
  });

  it('rejects a bid that is not 0/16/20/24/28', () => {
    expect('error' in parseTwentyNineCommand('bid 10')).toBe(true);
    expect('error' in parseTwentyNineCommand('bid 30')).toBe(true);
  });

  it('rejects a bid without a number', () => {
    expect('error' in parseTwentyNineCommand('bid')).toBe(true);
  });

  it('parses pass as bid 0', () => {
    expect(parseTwentyNineCommand('pass')).toEqual({ args: ['bid', { bid: 0 }] });
  });

  it('parses play (short and long)', () => {
    expect(parseTwentyNineCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseTwentyNineCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseTwentyNineCommand('p')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseTwentyNineCommand('n')).toEqual({ args: ['next'] });
    expect(parseTwentyNineCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseTwentyNineCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseTwentyNineCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseTwentyNineCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parseTwentyNineCommand('sd 9')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseTwentyNineCommand('h')).toEqual({ args: ['hint'] });
    expect(parseTwentyNineCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseTwentyNineCommand('l')).toEqual({ args: ['log'] });
    expect(parseTwentyNineCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseTwentyNineCommand('r')).toEqual({ args: ['reset'] });
    expect(parseTwentyNineCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseTwentyNineCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseTwentyNineCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(TWENTY_NINE_HELP.length).toBeGreaterThan(0);
    expect(TWENTY_NINE_HELP.some((line) => line.toLowerCase().includes('bid'))).toBe(true);
  });
});
