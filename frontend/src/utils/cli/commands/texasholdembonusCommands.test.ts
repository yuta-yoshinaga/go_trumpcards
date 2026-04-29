import { describe, expect, it } from 'vitest';
import { parseTexasholdembonusCommand, TEXASHOLDEMBONUS_HELP } from './texasholdembonusCommands';

describe('parseTexasholdembonusCommand', () => {
  it('parses bet with ante only', () => {
    expect(parseTexasholdembonusCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseTexasholdembonusCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('parses bet with ante and bonus side bet', () => {
    expect(parseTexasholdembonusCommand('b 100 10')).toEqual({ args: ['bet', 100, 10] });
    expect(parseTexasholdembonusCommand('bet 200 25')).toEqual({ args: ['bet', 200, 25] });
  });

  it('returns error for bet without amount', () => {
    expect('error' in parseTexasholdembonusCommand('b')).toBe(true);
    expect('error' in parseTexasholdembonusCommand('bet')).toBe(true);
  });

  it('returns error for non-numeric ante', () => {
    expect('error' in parseTexasholdembonusCommand('b abc')).toBe(true);
  });

  it('returns error for non-numeric bonus', () => {
    expect('error' in parseTexasholdembonusCommand('b 100 xyz')).toBe(true);
  });

  it('parses play', () => {
    expect(parseTexasholdembonusCommand('p')).toEqual({ args: ['play'] });
    expect(parseTexasholdembonusCommand('play')).toEqual({ args: ['play'] });
  });

  it('parses fold', () => {
    expect(parseTexasholdembonusCommand('f')).toEqual({ args: ['fold'] });
    expect(parseTexasholdembonusCommand('fold')).toEqual({ args: ['fold'] });
  });

  it('parses check', () => {
    expect(parseTexasholdembonusCommand('c')).toEqual({ args: ['check'] });
    expect(parseTexasholdembonusCommand('check')).toEqual({ args: ['check'] });
  });

  it('parses raise', () => {
    expect(parseTexasholdembonusCommand('ra')).toEqual({ args: ['raise'] });
    expect(parseTexasholdembonusCommand('raise')).toEqual({ args: ['raise'] });
  });

  it('parses log', () => {
    expect(parseTexasholdembonusCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseTexasholdembonusCommand('r')).toEqual({ args: ['reset'] });
    expect(parseTexasholdembonusCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parseTexasholdembonusCommand('xyz')).toBe(true);
  });

  it('suggests a similar command for typos', () => {
    const result = parseTexasholdembonusCommand('bett 100');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toMatch(/Did you mean/);
    }
  });
});

describe('TEXASHOLDEMBONUS_HELP', () => {
  it('lists all primary commands', () => {
    const joined = TEXASHOLDEMBONUS_HELP.join(' ');
    expect(joined).toMatch(/b /);
    expect(joined).toMatch(/p\/play/);
    expect(joined).toMatch(/f\/fold/);
    expect(joined).toMatch(/c\/check/);
    expect(joined).toMatch(/ra\/raise/);
    expect(joined).toMatch(/log/);
    expect(joined).toMatch(/r\/reset/);
  });
});
