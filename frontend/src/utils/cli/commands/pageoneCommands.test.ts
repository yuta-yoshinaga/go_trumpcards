import { describe, expect, it } from 'vitest';
import { PAGEONE_HELP, parsePageoneCommand } from './pageoneCommands';

describe('parsePageoneCommand', () => {
  it('parses play with index', () => {
    expect(parsePageoneCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parsePageoneCommand('play 5')).toEqual({ args: ['play', 5] });
  });

  it('returns error for play without index', () => {
    const result = parsePageoneCommand('p');
    expect('error' in result).toBe(true);
  });

  it('returns error for play with non-numeric index', () => {
    const result = parsePageoneCommand('p abc');
    expect('error' in result).toBe(true);
  });

  it('parses draw', () => {
    expect(parsePageoneCommand('d')).toEqual({ args: ['draw'] });
    expect(parsePageoneCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses declare', () => {
    expect(parsePageoneCommand('dc')).toEqual({ args: ['declare'] });
    expect(parsePageoneCommand('declare')).toEqual({ args: ['declare'] });
  });

  it('parses skip', () => {
    expect(parsePageoneCommand('sk')).toEqual({ args: ['skip'] });
    expect(parsePageoneCommand('skip')).toEqual({ args: ['skip'] });
  });

  it('parses nextround', () => {
    expect(parsePageoneCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parsePageoneCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses reset', () => {
    expect(parsePageoneCommand('r')).toEqual({ args: ['reset'] });
    expect(parsePageoneCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error with suggestion for typo', () => {
    const result = parsePageoneCommand('playy 3');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toContain('play');
    }
  });

  it('returns error for unknown command', () => {
    const result = parsePageoneCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('exposes help lines', () => {
    expect(PAGEONE_HELP.length).toBeGreaterThan(0);
  });
});
