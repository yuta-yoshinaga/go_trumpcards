import { describe, expect, it } from 'vitest';
import { CUARENTA_HELP, parseCuarentaCommand } from './cuarentaCommands';

describe('parseCuarentaCommand', () => {
  it('parses play (short and long) into a 0-based handIndex', () => {
    expect(parseCuarentaCommand('p 1')).toEqual({ args: ['play', { handIndex: 0 }] });
    expect(parseCuarentaCommand('play 3')).toEqual({ args: ['play', { handIndex: 2 }] });
  });

  it('rejects an invalid play argument', () => {
    expect('error' in parseCuarentaCommand('p')).toBe(true);
    expect('error' in parseCuarentaCommand('p 0')).toBe(true);
    expect('error' in parseCuarentaCommand('p x')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseCuarentaCommand('n')).toEqual({ args: ['next'] });
    expect(parseCuarentaCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseCuarentaCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
    expect('error' in parseCuarentaCommand('sd 9')).toBe(true);
    expect('error' in parseCuarentaCommand('sd')).toBe(true);
  });

  it('parses log and reset (short and long)', () => {
    expect(parseCuarentaCommand('l')).toEqual({ args: ['log'] });
    expect(parseCuarentaCommand('log')).toEqual({ args: ['log'] });
    expect(parseCuarentaCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCuarentaCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseCuarentaCommand('pla');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseCuarentaCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(CUARENTA_HELP.length).toBeGreaterThan(0);
    expect(CUARENTA_HELP.some((line) => line.toLowerCase().includes('play'))).toBe(true);
  });
});
