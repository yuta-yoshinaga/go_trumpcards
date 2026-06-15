import { describe, expect, it } from 'vitest';
import { PREFERENCE_HELP, parsePreferenceCommand } from './preferenceCommands';

describe('parsePreferenceCommand', () => {
  it('parses bid (short and long)', () => {
    expect(parsePreferenceCommand('b 1')).toEqual({ args: ['bid', { bid: 1 }] });
    expect(parsePreferenceCommand('bid 4')).toEqual({ args: ['bid', { bid: 4 }] });
  });

  it('rejects an out-of-range bid', () => {
    expect('error' in parsePreferenceCommand('bid 5')).toBe(true);
  });

  it('rejects a bid without a number', () => {
    expect('error' in parsePreferenceCommand('bid')).toBe(true);
  });

  it('parses pass as bid 0', () => {
    expect(parsePreferenceCommand('pass')).toEqual({ args: ['bid', { bid: 0 }] });
  });

  it('parses play (short and long)', () => {
    expect(parsePreferenceCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parsePreferenceCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parsePreferenceCommand('p')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parsePreferenceCommand('n')).toEqual({ args: ['next'] });
    expect(parsePreferenceCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parsePreferenceCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parsePreferenceCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parsePreferenceCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parsePreferenceCommand('sd 9')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parsePreferenceCommand('h')).toEqual({ args: ['hint'] });
    expect(parsePreferenceCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parsePreferenceCommand('l')).toEqual({ args: ['log'] });
    expect(parsePreferenceCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parsePreferenceCommand('r')).toEqual({ args: ['reset'] });
    expect(parsePreferenceCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parsePreferenceCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parsePreferenceCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(PREFERENCE_HELP.length).toBeGreaterThan(0);
    expect(PREFERENCE_HELP.some((line) => line.includes('Bid'))).toBe(true);
  });
});
