import { describe, expect, it } from 'vitest';
import { BOUILLOTTE_HELP, parseBouillotteCommand } from './bouillotteCommands';

describe('parseBouillotteCommand', () => {
  it('parses call (short and long) to bet call', () => {
    expect(parseBouillotteCommand('c')).toEqual({ args: ['bet', 'call'] });
    expect(parseBouillotteCommand('call')).toEqual({ args: ['bet', 'call'] });
  });

  it('parses raise aliases to bet raise', () => {
    expect(parseBouillotteCommand('ra')).toEqual({ args: ['bet', 'raise'] });
    expect(parseBouillotteCommand('raise')).toEqual({ args: ['bet', 'raise'] });
    expect(parseBouillotteCommand('vie')).toEqual({ args: ['bet', 'raise'] });
  });

  it('parses fold (short and long) to bet fold', () => {
    expect(parseBouillotteCommand('f')).toEqual({ args: ['bet', 'fold'] });
    expect(parseBouillotteCommand('fold')).toEqual({ args: ['bet', 'fold'] });
  });

  it('parses next aliases to nextround', () => {
    expect(parseBouillotteCommand('n')).toEqual({ args: ['nextround'] });
    expect(parseBouillotteCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseBouillotteCommand('next')).toEqual({ args: ['nextround'] });
    expect(parseBouillotteCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sp into a reset with playerCount config', () => {
    expect(parseBouillotteCommand('sp 3')).toEqual({ args: ['reset', undefined, { playerCount: 3 }] });
  });

  it('rejects an out-of-range sp', () => {
    expect('error' in parseBouillotteCommand('sp 2')).toBe(true);
    expect('error' in parseBouillotteCommand('sp 5')).toBe(true);
  });

  it('parses sa into a reset with ante config', () => {
    expect(parseBouillotteCommand('sa 25')).toEqual({ args: ['reset', undefined, { ante: 25 }] });
  });

  it('rejects an out-of-range sa', () => {
    expect('error' in parseBouillotteCommand('sa 0')).toBe(true);
  });

  it('parses sc into a reset with startingChips config', () => {
    expect(parseBouillotteCommand('sc 500')).toEqual({ args: ['reset', undefined, { startingChips: 500 }] });
  });

  it('rejects an out-of-range sc', () => {
    expect('error' in parseBouillotteCommand('sc 5')).toBe(true);
  });

  it('parses st into a reset with targetRounds config', () => {
    expect(parseBouillotteCommand('st 20')).toEqual({ args: ['reset', undefined, { targetRounds: 20 }] });
  });

  it('rejects an out-of-range st', () => {
    expect('error' in parseBouillotteCommand('st 0')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseBouillotteCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBouillotteCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseBouillotteCommand('l')).toEqual({ args: ['log'] });
    expect(parseBouillotteCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseBouillotteCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBouillotteCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseBouillotteCommand('cal');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseBouillotteCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(BOUILLOTTE_HELP.length).toBeGreaterThan(0);
    expect(BOUILLOTTE_HELP.some((line) => line.toLowerCase().includes('call'))).toBe(true);
  });
});
