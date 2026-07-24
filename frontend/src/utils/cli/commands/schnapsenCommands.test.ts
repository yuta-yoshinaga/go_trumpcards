import { describe, expect, it } from 'vitest';
import { parseSchnapsenCommand } from './schnapsenCommands';

describe('parseSchnapsenCommand', () => {
  it('parses play with a card index', () => {
    expect(parseSchnapsenCommand('play 2')).toEqual({ args: ['play', 2] });
    expect(parseSchnapsenCommand('p 0')).toEqual({ args: ['play', 0] });
  });

  it('errors on play without a valid index', () => {
    expect('error' in parseSchnapsenCommand('play')).toBe(true);
    expect('error' in parseSchnapsenCommand('p abc')).toBe(true);
  });

  it('parses marriage with a card index', () => {
    expect(parseSchnapsenCommand('marriage 3')).toEqual({ args: ['marriage', 3] });
    expect(parseSchnapsenCommand('m 1')).toEqual({ args: ['marriage', 1] });
  });

  it('errors on marriage without a valid index', () => {
    expect('error' in parseSchnapsenCommand('marriage')).toBe(true);
  });

  it('parses next, hint, log, and reset', () => {
    expect(parseSchnapsenCommand('next')).toEqual({ args: ['next'] });
    expect(parseSchnapsenCommand('n')).toEqual({ args: ['next'] });
    expect(parseSchnapsenCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseSchnapsenCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSchnapsenCommand('log')).toEqual({ args: ['log'] });
    expect(parseSchnapsenCommand('l')).toEqual({ args: ['log'] });
    expect(parseSchnapsenCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseSchnapsenCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for a close typo', () => {
    const result = parseSchnapsenCommand('marraige 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('marriage');
  });

  it('errors on an unknown command', () => {
    expect('error' in parseSchnapsenCommand('zzz')).toBe(true);
  });
});
