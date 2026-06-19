import { describe, expect, it } from 'vitest';
import { PISHTI_HELP, parsePishtiCommand } from './pishtiCommands';

describe('parsePishtiCommand', () => {
  it('parses play (short and long) into a 0-based handIndex', () => {
    expect(parsePishtiCommand('p 1')).toEqual({ args: ['play', { handIndex: 0 }] });
    expect(parsePishtiCommand('play 3')).toEqual({ args: ['play', { handIndex: 2 }] });
  });

  it('rejects an invalid play argument', () => {
    expect('error' in parsePishtiCommand('p')).toBe(true);
    expect('error' in parsePishtiCommand('p 0')).toBe(true);
    expect('error' in parsePishtiCommand('p x')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parsePishtiCommand('n')).toEqual({ args: ['next'] });
    expect(parsePishtiCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parsePishtiCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
    expect('error' in parsePishtiCommand('sd 9')).toBe(true);
    expect('error' in parsePishtiCommand('sd')).toBe(true);
  });

  it('parses sp into a reset with player-count config', () => {
    expect(parsePishtiCommand('sp 3')).toEqual({ args: ['reset', { config: { playerCnt: 3 } }] });
    expect('error' in parsePishtiCommand('sp 5')).toBe(true);
    expect('error' in parsePishtiCommand('sp 1')).toBe(true);
  });

  it('parses log and reset (short and long)', () => {
    expect(parsePishtiCommand('l')).toEqual({ args: ['log'] });
    expect(parsePishtiCommand('log')).toEqual({ args: ['log'] });
    expect(parsePishtiCommand('r')).toEqual({ args: ['reset'] });
    expect(parsePishtiCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parsePishtiCommand('pla');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parsePishtiCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(PISHTI_HELP.length).toBeGreaterThan(0);
    expect(PISHTI_HELP.some((line) => line.toLowerCase().includes('play'))).toBe(true);
  });
});
