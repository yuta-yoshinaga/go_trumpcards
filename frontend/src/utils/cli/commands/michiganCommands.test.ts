import { describe, expect, it } from 'vitest';
import { MICHIGAN_HELP, parseMichiganCommand } from './michiganCommands';

describe('parseMichiganCommand', () => {
  it('parses bet (short and long) with four boodle amounts', () => {
    expect(parseMichiganCommand('b 2 2 2 2')).toEqual({ args: ['bet', [2, 2, 2, 2]] });
    expect(parseMichiganCommand('bet 1 3 0 4')).toEqual({ args: ['bet', [1, 3, 0, 4]] });
  });

  it('rejects a bet without four amounts', () => {
    expect('error' in parseMichiganCommand('b 2 2')).toBe(true);
  });

  it('rejects a bet with a negative or non-numeric amount', () => {
    expect('error' in parseMichiganCommand('b 2 2 2 -1')).toBe(true);
    expect('error' in parseMichiganCommand('b a b c d')).toBe(true);
  });

  it('parses play (short and long) to a play with card index', () => {
    expect(parseMichiganCommand('p 0')).toEqual({ args: ['play', undefined, 0] });
    expect(parseMichiganCommand('play 3')).toEqual({ args: ['play', undefined, 3] });
  });

  it('rejects a play without a valid index', () => {
    expect('error' in parseMichiganCommand('p')).toBe(true);
    expect('error' in parseMichiganCommand('p x')).toBe(true);
  });

  it('parses next aliases to nextround', () => {
    expect(parseMichiganCommand('n')).toEqual({ args: ['nextround'] });
    expect(parseMichiganCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseMichiganCommand('next')).toEqual({ args: ['nextround'] });
    expect(parseMichiganCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sp into a reset with playerCount config', () => {
    expect(parseMichiganCommand('sp 6')).toEqual({ args: ['reset', undefined, undefined, { playerCount: 6 }] });
  });

  it('rejects an out-of-range sp', () => {
    expect('error' in parseMichiganCommand('sp 2')).toBe(true);
    expect('error' in parseMichiganCommand('sp 9')).toBe(true);
  });

  it('parses sa into a reset with ante config', () => {
    expect(parseMichiganCommand('sa 12')).toEqual({ args: ['reset', undefined, undefined, { ante: 12 }] });
  });

  it('rejects an out-of-range sa', () => {
    expect('error' in parseMichiganCommand('sa 3')).toBe(true);
  });

  it('parses sc into a reset with startingChips config', () => {
    expect(parseMichiganCommand('sc 500')).toEqual({ args: ['reset', undefined, undefined, { startingChips: 500 }] });
  });

  it('rejects an out-of-range sc', () => {
    expect('error' in parseMichiganCommand('sc 5')).toBe(true);
  });

  it('parses st into a reset with targetRounds config', () => {
    expect(parseMichiganCommand('st 20')).toEqual({ args: ['reset', undefined, undefined, { targetRounds: 20 }] });
  });

  it('rejects an out-of-range st', () => {
    expect('error' in parseMichiganCommand('st 0')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseMichiganCommand('h')).toEqual({ args: ['hint'] });
    expect(parseMichiganCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseMichiganCommand('l')).toEqual({ args: ['log'] });
    expect(parseMichiganCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseMichiganCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMichiganCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseMichiganCommand('ply');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseMichiganCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(MICHIGAN_HELP.length).toBeGreaterThan(0);
    expect(MICHIGAN_HELP.some((line) => line.toLowerCase().includes('play'))).toBe(true);
  });
});
