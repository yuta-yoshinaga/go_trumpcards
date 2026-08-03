import { describe, expect, it } from 'vitest';
import { PREFERENCE_HELP, parseViraCommand } from './viraCommands';

describe('parseViraCommand', () => {
  it('parses bid (short and long)', () => {
    expect(parseViraCommand('b 1')).toEqual({ args: ['bid', { bid: 1 }] });
    expect(parseViraCommand('bid 4')).toEqual({ args: ['bid', { bid: 4 }] });
  });

  it('rejects an out-of-range bid', () => {
    expect('error' in parseViraCommand('bid 5')).toBe(true);
  });

  it('rejects a bid without a number', () => {
    expect('error' in parseViraCommand('bid')).toBe(true);
  });

  it('parses pass as bid 0', () => {
    expect(parseViraCommand('pass')).toEqual({ args: ['bid', { bid: 0 }] });
  });

  it('parses play (short and long)', () => {
    expect(parseViraCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseViraCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseViraCommand('p')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseViraCommand('n')).toEqual({ args: ['next'] });
    expect(parseViraCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseViraCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseViraCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseViraCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parseViraCommand('sd 9')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseViraCommand('h')).toEqual({ args: ['hint'] });
    expect(parseViraCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseViraCommand('l')).toEqual({ args: ['log'] });
    expect(parseViraCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseViraCommand('r')).toEqual({ args: ['reset'] });
    expect(parseViraCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseViraCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseViraCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(PREFERENCE_HELP.length).toBeGreaterThan(0);
    expect(PREFERENCE_HELP.some((line) => line.includes('Bid'))).toBe(true);
  });
});
