import { describe, expect, it } from 'vitest';
import { parseRistikontraCommand, RISTIKONTRA_HELP } from './ristikontraCommands';

describe('parseRistikontraCommand', () => {
  it('parses play (short and long) into a 0-based handIndex', () => {
    expect(parseRistikontraCommand('p 1')).toEqual({ args: ['play', { handIndex: 0 }] });
    expect(parseRistikontraCommand('play 3')).toEqual({ args: ['play', { handIndex: 2 }] });
  });

  it('rejects an invalid play argument', () => {
    expect('error' in parseRistikontraCommand('p')).toBe(true);
    expect('error' in parseRistikontraCommand('p 0')).toBe(true);
    expect('error' in parseRistikontraCommand('p x')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseRistikontraCommand('n')).toEqual({ args: ['next'] });
    expect(parseRistikontraCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses sd into a reset with difficulty config', () => {
    expect(parseRistikontraCommand('sd 2')).toEqual({ args: ['reset', { config: { cpuDifficulty: 2 } }] });
    expect('error' in parseRistikontraCommand('sd 9')).toBe(true);
    expect('error' in parseRistikontraCommand('sd')).toBe(true);
  });

  it('parses sp into a reset with player-count config', () => {
    expect(parseRistikontraCommand('sp 3')).toEqual({ args: ['reset', { config: { playerCnt: 3 } }] });
    expect('error' in parseRistikontraCommand('sp 5')).toBe(true);
    expect('error' in parseRistikontraCommand('sp 1')).toBe(true);
  });

  it('parses log and reset (short and long)', () => {
    expect(parseRistikontraCommand('l')).toEqual({ args: ['log'] });
    expect(parseRistikontraCommand('log')).toEqual({ args: ['log'] });
    expect(parseRistikontraCommand('r')).toEqual({ args: ['reset'] });
    expect(parseRistikontraCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseRistikontraCommand('pla');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseRistikontraCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(RISTIKONTRA_HELP.length).toBeGreaterThan(0);
    expect(RISTIKONTRA_HELP.some((line) => line.toLowerCase().includes('play'))).toBe(true);
  });
});
