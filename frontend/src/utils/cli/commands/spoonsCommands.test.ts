import { describe, expect, it } from 'vitest';
import { parseSpoonsCommand, SPOONS_HELP } from './spoonsCommands';

describe('parseSpoonsCommand', () => {
  it('parses pass with an index (short and long)', () => {
    expect(parseSpoonsCommand('p 2')).toEqual({ args: ['pass', { cardIndex: 2 }] });
    expect(parseSpoonsCommand('pass 0')).toEqual({ args: ['pass', { cardIndex: 0 }] });
  });

  it('returns error for pass without a valid index', () => {
    expect('error' in parseSpoonsCommand('p')).toBe(true);
    expect('error' in parseSpoonsCommand('p -1')).toBe(true);
  });

  it('parses grab (short and long)', () => {
    expect(parseSpoonsCommand('g')).toEqual({ args: ['grab'] });
    expect(parseSpoonsCommand('grab')).toEqual({ args: ['grab'] });
  });

  it('parses next (short and long)', () => {
    expect(parseSpoonsCommand('n')).toEqual({ args: ['next'] });
    expect(parseSpoonsCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseSpoonsCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
  });

  it('rejects an out-of-range sd', () => {
    expect('error' in parseSpoonsCommand('sd 9')).toBe(true);
  });

  it('parses log (short and long)', () => {
    expect(parseSpoonsCommand('l')).toEqual({ args: ['log'] });
    expect(parseSpoonsCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseSpoonsCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSpoonsCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseSpoonsCommand('gra');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseSpoonsCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(SPOONS_HELP.length).toBeGreaterThan(0);
    expect(SPOONS_HELP.some((line) => line.toLowerCase().includes('grab'))).toBe(true);
    expect(SPOONS_HELP.some((line) => line.toLowerCase().includes('pass'))).toBe(true);
  });
});
