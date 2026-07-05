import { describe, expect, it } from 'vitest';
import { ANACONDA_HELP, parseAnacondaCommand } from './anacondaCommands';

describe('parseAnacondaCommand', () => {
  it('parses pass (short and long) with card indices', () => {
    expect(parseAnacondaCommand('p 0 1 2')).toEqual({ args: ['pass', [0, 1, 2]] });
    expect(parseAnacondaCommand('pass 4 5')).toEqual({ args: ['pass', [4, 5]] });
    expect(parseAnacondaCommand('p 6')).toEqual({ args: ['pass', [6]] });
  });

  it('rejects a pass with no indices or too many', () => {
    expect('error' in parseAnacondaCommand('p')).toBe(true);
    expect('error' in parseAnacondaCommand('p 0 1 2 3')).toBe(true);
    expect('error' in parseAnacondaCommand('p x')).toBe(true);
  });

  it('parses keep (short and long) with exactly 5 indices', () => {
    expect(parseAnacondaCommand('k 0 1 2 3 4')).toEqual({ args: ['keep', [0, 1, 2, 3, 4]] });
    expect(parseAnacondaCommand('keep 2 3 4 5 6')).toEqual({ args: ['keep', [2, 3, 4, 5, 6]] });
  });

  it('rejects a keep that is not exactly 5 indices', () => {
    expect('error' in parseAnacondaCommand('k 0 1 2 3')).toBe(true);
    expect('error' in parseAnacondaCommand('k 0 1 2 3 4 5')).toBe(true);
  });

  it('parses call/check aliases to a call bet', () => {
    expect(parseAnacondaCommand('c')).toEqual({ args: ['bet', undefined, 'call'] });
    expect(parseAnacondaCommand('call')).toEqual({ args: ['bet', undefined, 'call'] });
    expect(parseAnacondaCommand('check')).toEqual({ args: ['bet', undefined, 'call'] });
  });

  it('parses raise aliases to a raise bet', () => {
    expect(parseAnacondaCommand('ra')).toEqual({ args: ['bet', undefined, 'raise'] });
    expect(parseAnacondaCommand('raise')).toEqual({ args: ['bet', undefined, 'raise'] });
  });

  it('parses fold aliases to a fold bet', () => {
    expect(parseAnacondaCommand('f')).toEqual({ args: ['bet', undefined, 'fold'] });
    expect(parseAnacondaCommand('fold')).toEqual({ args: ['bet', undefined, 'fold'] });
  });

  it('parses next aliases to nextround', () => {
    expect(parseAnacondaCommand('n')).toEqual({ args: ['nextround'] });
    expect(parseAnacondaCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseAnacondaCommand('next')).toEqual({ args: ['nextround'] });
    expect(parseAnacondaCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses sp into a reset with playerCount config', () => {
    expect(parseAnacondaCommand('sp 6')).toEqual({ args: ['reset', undefined, undefined, { playerCount: 6 }] });
  });

  it('rejects an out-of-range sp', () => {
    expect('error' in parseAnacondaCommand('sp 2')).toBe(true);
    expect('error' in parseAnacondaCommand('sp 8')).toBe(true);
  });

  it('parses sa into a reset with ante config', () => {
    expect(parseAnacondaCommand('sa 25')).toEqual({ args: ['reset', undefined, undefined, { ante: 25 }] });
  });

  it('rejects an out-of-range sa', () => {
    expect('error' in parseAnacondaCommand('sa 0')).toBe(true);
  });

  it('parses sc into a reset with startingChips config', () => {
    expect(parseAnacondaCommand('sc 500')).toEqual({ args: ['reset', undefined, undefined, { startingChips: 500 }] });
  });

  it('rejects an out-of-range sc', () => {
    expect('error' in parseAnacondaCommand('sc 5')).toBe(true);
  });

  it('parses st into a reset with targetRounds config', () => {
    expect(parseAnacondaCommand('st 20')).toEqual({ args: ['reset', undefined, undefined, { targetRounds: 20 }] });
  });

  it('rejects an out-of-range st', () => {
    expect('error' in parseAnacondaCommand('st 0')).toBe(true);
  });

  it('parses hint (short and long)', () => {
    expect(parseAnacondaCommand('h')).toEqual({ args: ['hint'] });
    expect(parseAnacondaCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log (short and long)', () => {
    expect(parseAnacondaCommand('l')).toEqual({ args: ['log'] });
    expect(parseAnacondaCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseAnacondaCommand('r')).toEqual({ args: ['reset'] });
    expect(parseAnacondaCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseAnacondaCommand('passs');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseAnacondaCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(ANACONDA_HELP.length).toBeGreaterThan(0);
    expect(ANACONDA_HELP.some((line) => line.toLowerCase().includes('pass'))).toBe(true);
  });
});
