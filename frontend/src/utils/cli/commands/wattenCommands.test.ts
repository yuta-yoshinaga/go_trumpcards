import { describe, expect, it } from 'vitest';
import { parseWattenCommand, WATTEN_HELP } from './wattenCommands';

describe('parseWattenCommand', () => {
  it('parses declare with a rank letter and suit letter', () => {
    expect(parseWattenCommand('declare k h')).toEqual({ args: ['declare', 13, 3] });
    expect(parseWattenCommand('d a s')).toEqual({ args: ['declare', 1, 1] });
    expect(parseWattenCommand('d j c')).toEqual({ args: ['declare', 11, 2] });
  });

  it('parses declare with a numeric rank', () => {
    expect(parseWattenCommand('d 7 d')).toEqual({ args: ['declare', 7, 4] });
    expect(parseWattenCommand('d 10 h')).toEqual({ args: ['declare', 10, 3] });
  });

  it('returns error for declare without a valid rank or suit', () => {
    expect('error' in parseWattenCommand('d')).toBe(true);
    expect('error' in parseWattenCommand('d k')).toBe(true);
    expect('error' in parseWattenCommand('d x h')).toBe(true);
    expect('error' in parseWattenCommand('d k z')).toBe(true);
  });

  it('parses play (short and long)', () => {
    expect(parseWattenCommand('p 2')).toEqual({ args: ['play', undefined, undefined, 2] });
    expect(parseWattenCommand('play 0')).toEqual({ args: ['play', undefined, undefined, 0] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseWattenCommand('p')).toBe(true);
  });

  it('parses raise (short and long)', () => {
    expect(parseWattenCommand('rz')).toEqual({ args: ['raise'] });
    expect(parseWattenCommand('raise')).toEqual({ args: ['raise'] });
  });

  it('parses hold and fold as respond commands', () => {
    expect(parseWattenCommand('hold')).toEqual({ args: ['respond', undefined, undefined, undefined, true] });
    expect(parseWattenCommand('fold')).toEqual({ args: ['respond', undefined, undefined, undefined, false] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseWattenCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseWattenCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseWattenCommand('h')).toEqual({ args: ['hint'] });
    expect(parseWattenCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseWattenCommand('r')).toEqual({ args: ['reset'] });
    expect(parseWattenCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseWattenCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseWattenCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(WATTEN_HELP.length).toBeGreaterThan(0);
    expect(WATTEN_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
