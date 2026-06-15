import { describe, expect, it } from 'vitest';
import { parseSoloWhistCommand, SOLO_WHIST_HELP } from './soloWhistCommands';

describe('parseSoloWhistCommand', () => {
  it('parses bid (short and long)', () => {
    expect(parseSoloWhistCommand('b 1')).toEqual({ args: ['bid', { bid: 1 }] });
    expect(parseSoloWhistCommand('bid 3')).toEqual({ args: ['bid', { bid: 3 }] });
  });

  it('rejects an out-of-range bid', () => {
    const result = parseSoloWhistCommand('bid 4');
    expect('error' in result).toBe(true);
  });

  it('rejects a bid without a number', () => {
    const result = parseSoloWhistCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('parses pass as bid 0', () => {
    expect(parseSoloWhistCommand('pass')).toEqual({ args: ['bid', { bid: 0 }] });
  });

  it('parses play (short and long)', () => {
    expect(parseSoloWhistCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseSoloWhistCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseSoloWhistCommand('p')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseSoloWhistCommand('n')).toEqual({ args: ['next'] });
    expect(parseSoloWhistCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseSoloWhistCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseSoloWhistCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseSoloWhistCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parseSoloWhistCommand('sd 9')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseSoloWhistCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSoloWhistCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseSoloWhistCommand('l')).toEqual({ args: ['log'] });
    expect(parseSoloWhistCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseSoloWhistCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSoloWhistCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseSoloWhistCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseSoloWhistCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(SOLO_WHIST_HELP.length).toBeGreaterThan(0);
    expect(SOLO_WHIST_HELP.some((line) => line.includes('Bid'))).toBe(true);
  });
});
