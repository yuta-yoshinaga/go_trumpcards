import { describe, expect, it } from 'vitest';
import { NAP_HELP, parseNapCommand } from './napCommands';

describe('parseNapCommand', () => {
  it('parses bid (short and long)', () => {
    expect(parseNapCommand('b 2')).toEqual({ args: ['bid', { bid: 2 }] });
    expect(parseNapCommand('bid 5')).toEqual({ args: ['bid', { bid: 5 }] });
  });

  it('rejects an out-of-range / invalid bid', () => {
    expect('error' in parseNapCommand('bid 1')).toBe(true);
    expect('error' in parseNapCommand('bid 6')).toBe(true);
  });

  it('rejects a bid without a number', () => {
    expect('error' in parseNapCommand('bid')).toBe(true);
  });

  it('parses pass as bid 0', () => {
    expect(parseNapCommand('pass')).toEqual({ args: ['bid', { bid: 0 }] });
  });

  it('parses play (short and long)', () => {
    expect(parseNapCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseNapCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseNapCommand('p')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseNapCommand('n')).toEqual({ args: ['next'] });
    expect(parseNapCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseNapCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseNapCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseNapCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parseNapCommand('sd 9')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseNapCommand('h')).toEqual({ args: ['hint'] });
    expect(parseNapCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseNapCommand('l')).toEqual({ args: ['log'] });
    expect(parseNapCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseNapCommand('r')).toEqual({ args: ['reset'] });
    expect(parseNapCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseNapCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseNapCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(NAP_HELP.length).toBeGreaterThan(0);
    expect(NAP_HELP.some((line) => line.includes('Bid'))).toBe(true);
  });
});
