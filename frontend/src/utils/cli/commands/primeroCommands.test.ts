import { describe, expect, it } from 'vitest';
import { PRIMERO_HELP, parsePrimeroCommand } from './primeroCommands';

describe('parsePrimeroCommand', () => {
  it('parses call (short and long) to bet call', () => {
    expect(parsePrimeroCommand('c')).toEqual({ args: ['bet', 'call'] });
    expect(parsePrimeroCommand('call')).toEqual({ args: ['bet', 'call'] });
  });

  it('parses raise aliases to bet raise', () => {
    expect(parsePrimeroCommand('ra')).toEqual({ args: ['bet', 'raise'] });
    expect(parsePrimeroCommand('raise')).toEqual({ args: ['bet', 'raise'] });
    expect(parsePrimeroCommand('vie')).toEqual({ args: ['bet', 'raise'] });
  });

  it('parses fold (short and long) to bet fold', () => {
    expect(parsePrimeroCommand('f')).toEqual({ args: ['bet', 'fold'] });
    expect(parsePrimeroCommand('fold')).toEqual({ args: ['bet', 'fold'] });
  });

  it('parses next aliases to nextround', () => {
    expect(parsePrimeroCommand('n')).toEqual({ args: ['nextround'] });
    expect(parsePrimeroCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parsePrimeroCommand('next')).toEqual({ args: ['nextround'] });
    expect(parsePrimeroCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sp into a reset with playerCount config', () => {
    expect(parsePrimeroCommand('sp 3')).toEqual({ args: ['reset', undefined, { playerCount: 3 }] });
    expect(parsePrimeroCommand('sp 6')).toEqual({ args: ['reset', undefined, { playerCount: 6 }] });
  });

  it('rejects an out-of-range sp', () => {
    expect('error' in parsePrimeroCommand('sp 1')).toBe(true);
    expect('error' in parsePrimeroCommand('sp 7')).toBe(true);
  });

  it('parses sa into a reset with ante config', () => {
    expect(parsePrimeroCommand('sa 25')).toEqual({ args: ['reset', undefined, { ante: 25 }] });
  });

  it('rejects an out-of-range sa', () => {
    expect('error' in parsePrimeroCommand('sa 0')).toBe(true);
  });

  it('parses sc into a reset with startingChips config', () => {
    expect(parsePrimeroCommand('sc 500')).toEqual({ args: ['reset', undefined, { startingChips: 500 }] });
  });

  it('rejects an out-of-range sc', () => {
    expect('error' in parsePrimeroCommand('sc 5')).toBe(true);
  });

  it('parses st into a reset with targetRounds config', () => {
    expect(parsePrimeroCommand('st 20')).toEqual({ args: ['reset', undefined, { targetRounds: 20 }] });
  });

  it('rejects an out-of-range st', () => {
    expect('error' in parsePrimeroCommand('st 0')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parsePrimeroCommand('h')).toEqual({ args: ['hint'] });
    expect(parsePrimeroCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parsePrimeroCommand('l')).toEqual({ args: ['log'] });
    expect(parsePrimeroCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parsePrimeroCommand('r')).toEqual({ args: ['reset'] });
    expect(parsePrimeroCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parsePrimeroCommand('cal');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parsePrimeroCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(PRIMERO_HELP.length).toBeGreaterThan(0);
    expect(PRIMERO_HELP.some((line) => line.toLowerCase().includes('call'))).toBe(true);
  });
});
