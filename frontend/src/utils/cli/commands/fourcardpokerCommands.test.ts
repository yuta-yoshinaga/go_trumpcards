import { describe, expect, it } from 'vitest';
import { parseFourCardPokerCommand } from './fourcardpokerCommands';

describe('parseFourCardPokerCommand', () => {
  it('parses bet with an ante only (acesUp defaults to 0)', () => {
    expect(parseFourCardPokerCommand('bet 100')).toEqual({ args: ['bet', 100, 0] });
    expect(parseFourCardPokerCommand('b 50')).toEqual({ args: ['bet', 50, 0] });
  });

  it('parses bet with an ante and an aces-up side bet', () => {
    expect(parseFourCardPokerCommand('bet 100 25')).toEqual({ args: ['bet', 100, 25] });
  });

  it('errors on bet without a valid ante', () => {
    expect('error' in parseFourCardPokerCommand('bet')).toBe(true);
    expect('error' in parseFourCardPokerCommand('bet abc')).toBe(true);
    expect('error' in parseFourCardPokerCommand('bet 100 xyz')).toBe(true);
  });

  it('parses play with a multiplier of 1-3', () => {
    expect(parseFourCardPokerCommand('play 1')).toEqual({ args: ['play', undefined, undefined, 1] });
    expect(parseFourCardPokerCommand('p 3')).toEqual({ args: ['play', undefined, undefined, 3] });
  });

  it('errors on an out-of-range or missing play multiplier', () => {
    expect('error' in parseFourCardPokerCommand('play')).toBe(true);
    expect('error' in parseFourCardPokerCommand('play 0')).toBe(true);
    expect('error' in parseFourCardPokerCommand('play 4')).toBe(true);
  });

  it('parses fold, log, and reset', () => {
    expect(parseFourCardPokerCommand('fold')).toEqual({ args: ['fold'] });
    expect(parseFourCardPokerCommand('f')).toEqual({ args: ['fold'] });
    expect(parseFourCardPokerCommand('log')).toEqual({ args: ['log'] });
    expect(parseFourCardPokerCommand('l')).toEqual({ args: ['log'] });
    expect(parseFourCardPokerCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseFourCardPokerCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for a close typo', () => {
    const result = parseFourCardPokerCommand('fol');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('fold');
  });

  it('errors on an unknown command', () => {
    expect('error' in parseFourCardPokerCommand('zzz')).toBe(true);
  });
});
