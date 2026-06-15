import { describe, expect, it } from 'vitest';
import { FORTY_FIVES_HELP, parseFortyFivesCommand } from './fortyFivesCommands';

describe('parseFortyFivesCommand', () => {
  it('parses bid (short and long)', () => {
    expect(parseFortyFivesCommand('b 15')).toEqual({ args: ['bid', { bid: 15 }] });
    expect(parseFortyFivesCommand('bid 25')).toEqual({ args: ['bid', { bid: 25 }] });
  });

  it('rejects a bid that is not 0/15/20/25', () => {
    expect('error' in parseFortyFivesCommand('bid 10')).toBe(true);
    expect('error' in parseFortyFivesCommand('bid 30')).toBe(true);
  });

  it('rejects a bid without a number', () => {
    expect('error' in parseFortyFivesCommand('bid')).toBe(true);
  });

  it('parses pass as bid 0', () => {
    expect(parseFortyFivesCommand('pass')).toEqual({ args: ['bid', { bid: 0 }] });
  });

  it('parses play (short and long)', () => {
    expect(parseFortyFivesCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseFortyFivesCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseFortyFivesCommand('p')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseFortyFivesCommand('n')).toEqual({ args: ['next'] });
    expect(parseFortyFivesCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseFortyFivesCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseFortyFivesCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseFortyFivesCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parseFortyFivesCommand('sd 9')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseFortyFivesCommand('h')).toEqual({ args: ['hint'] });
    expect(parseFortyFivesCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseFortyFivesCommand('l')).toEqual({ args: ['log'] });
    expect(parseFortyFivesCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseFortyFivesCommand('r')).toEqual({ args: ['reset'] });
    expect(parseFortyFivesCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseFortyFivesCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseFortyFivesCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(FORTY_FIVES_HELP.length).toBeGreaterThan(0);
    expect(FORTY_FIVES_HELP.some((line) => line.toLowerCase().includes('bid'))).toBe(true);
  });
});
