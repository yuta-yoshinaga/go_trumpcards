import { describe, expect, it } from 'vitest';
import { parseOasispokerCommand } from './oasispokerCommands';

describe('parseOasispokerCommand', () => {
  it('parses bet with amount', () => {
    expect(parseOasispokerCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseOasispokerCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('parses bet with amount and jackpot side bet', () => {
    expect(parseOasispokerCommand('b 100 10')).toEqual({ args: ['bet', 100, 10] });
  });

  it('returns error for bet without amount', () => {
    const result = parseOasispokerCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses exchange with indices', () => {
    expect(parseOasispokerCommand('e 0 2 4')).toEqual({
      args: ['exchange', undefined, undefined, [0, 2, 4]],
    });
    expect(parseOasispokerCommand('exchange 1')).toEqual({
      args: ['exchange', undefined, undefined, [1]],
    });
  });

  it('parses exchange with no indices (stand-equivalent)', () => {
    expect(parseOasispokerCommand('e')).toEqual({
      args: ['exchange', undefined, undefined, []],
    });
  });

  it('returns error for out-of-range exchange index', () => {
    expect('error' in parseOasispokerCommand('e 5')).toBe(true);
    expect('error' in parseOasispokerCommand('e -1')).toBe(true);
  });

  it('returns error for non-numeric exchange index', () => {
    expect('error' in parseOasispokerCommand('e abc')).toBe(true);
  });

  it('parses stand', () => {
    expect(parseOasispokerCommand('s')).toEqual({ args: ['stand'] });
    expect(parseOasispokerCommand('stand')).toEqual({ args: ['stand'] });
  });

  it('parses play', () => {
    expect(parseOasispokerCommand('p')).toEqual({ args: ['play'] });
    expect(parseOasispokerCommand('play')).toEqual({ args: ['play'] });
  });

  it('parses fold', () => {
    expect(parseOasispokerCommand('f')).toEqual({ args: ['fold'] });
  });

  it('parses log', () => {
    expect(parseOasispokerCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseOasispokerCommand('r')).toEqual({ args: ['reset'] });
    expect(parseOasispokerCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseOasispokerCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
