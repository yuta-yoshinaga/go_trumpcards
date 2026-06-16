import { describe, expect, it } from 'vitest';
import { BEZIQUE_HELP, parseBeziqueCommand } from './beziqueCommands';

describe('parseBeziqueCommand', () => {
  it('parses play (short and long)', () => {
    expect(parseBeziqueCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseBeziqueCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseBeziqueCommand('p')).toBe(true);
  });

  it('parses meld (short and long)', () => {
    expect(parseBeziqueCommand('m 1')).toEqual({ args: ['meld', { meldIndex: 1 }] });
    expect(parseBeziqueCommand('meld 0')).toEqual({ args: ['meld', { meldIndex: 0 }] });
  });

  it('returns error for meld without index', () => {
    expect('error' in parseBeziqueCommand('m')).toBe(true);
  });

  it('parses skip (short and long)', () => {
    expect(parseBeziqueCommand('s')).toEqual({ args: ['skip'] });
    expect(parseBeziqueCommand('skip')).toEqual({ args: ['skip'] });
  });

  it('parses next (short and long)', () => {
    expect(parseBeziqueCommand('n')).toEqual({ args: ['next'] });
    expect(parseBeziqueCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseBeziqueCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parseBeziqueCommand('sd 9')).toBe(true);
  });

  it('parses st into a reset with target-score config', () => {
    expect(parseBeziqueCommand('st 500')).toEqual({ args: ['reset', { config: { targetScore: 500 } }] });
  });

  it('rejects an out-of-range st', () => {
    expect('error' in parseBeziqueCommand('st 0')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseBeziqueCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBeziqueCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseBeziqueCommand('l')).toEqual({ args: ['log'] });
    expect(parseBeziqueCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseBeziqueCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBeziqueCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseBeziqueCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseBeziqueCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(BEZIQUE_HELP.length).toBeGreaterThan(0);
    expect(BEZIQUE_HELP.some((line) => line.toLowerCase().includes('meld'))).toBe(true);
  });
});
