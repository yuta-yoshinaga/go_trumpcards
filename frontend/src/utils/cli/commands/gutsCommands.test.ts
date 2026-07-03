import { describe, expect, it } from 'vitest';
import { GUTS_HELP, parseGutsCommand } from './gutsCommands';

describe('parseGutsCommand', () => {
  it('parses in (short and long) to declare 1', () => {
    expect(parseGutsCommand('i')).toEqual({ args: ['declare', 1] });
    expect(parseGutsCommand('in')).toEqual({ args: ['declare', 1] });
  });

  it('parses out (short and long) to declare 0', () => {
    expect(parseGutsCommand('o')).toEqual({ args: ['declare', 0] });
    expect(parseGutsCommand('out')).toEqual({ args: ['declare', 0] });
  });

  it('parses next aliases to nextround', () => {
    expect(parseGutsCommand('n')).toEqual({ args: ['nextround'] });
    expect(parseGutsCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseGutsCommand('next')).toEqual({ args: ['nextround'] });
    expect(parseGutsCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sp into a reset with playerCount config', () => {
    expect(parseGutsCommand('sp 6')).toEqual({ args: ['reset', undefined, { playerCount: 6 }] });
  });

  it('rejects an out-of-range sp', () => {
    expect('error' in parseGutsCommand('sp 1')).toBe(true);
    expect('error' in parseGutsCommand('sp 8')).toBe(true);
  });

  it('parses sa into a reset with ante config', () => {
    expect(parseGutsCommand('sa 25')).toEqual({ args: ['reset', undefined, { ante: 25 }] });
  });

  it('rejects an out-of-range sa', () => {
    expect('error' in parseGutsCommand('sa 0')).toBe(true);
  });

  it('parses sc into a reset with startingChips config', () => {
    expect(parseGutsCommand('sc 500')).toEqual({ args: ['reset', undefined, { startingChips: 500 }] });
  });

  it('rejects an out-of-range sc', () => {
    expect('error' in parseGutsCommand('sc 5')).toBe(true);
  });

  it('parses st into a reset with targetRounds config', () => {
    expect(parseGutsCommand('st 20')).toEqual({ args: ['reset', undefined, { targetRounds: 20 }] });
  });

  it('rejects an out-of-range st', () => {
    expect('error' in parseGutsCommand('st 0')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseGutsCommand('h')).toEqual({ args: ['hint'] });
    expect(parseGutsCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseGutsCommand('l')).toEqual({ args: ['log'] });
    expect(parseGutsCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseGutsCommand('r')).toEqual({ args: ['reset'] });
    expect(parseGutsCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseGutsCommand('inn');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseGutsCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(GUTS_HELP.length).toBeGreaterThan(0);
    expect(GUTS_HELP.some((line) => line.toLowerCase().includes('in'))).toBe(true);
  });
});
