import { describe, expect, it } from 'vitest';
import { parseDeuceswildCommand } from './deuceswildCommands';

describe('parseDeuceswildCommand', () => {
  it('parses bet with amount', () => {
    expect(parseDeuceswildCommand('b 100')).toEqual({ args: ['bet', 100, undefined] });
    expect(parseDeuceswildCommand('bet 50')).toEqual({ args: ['bet', 50, undefined] });
  });

  it('returns error for bet without amount', () => {
    const result = parseDeuceswildCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses hold with indices', () => {
    expect(parseDeuceswildCommand('hold 0 2 4')).toEqual({ args: ['hold', undefined, [0, 2, 4]] });
  });

  it('parses hold without indices as empty', () => {
    expect(parseDeuceswildCommand('hold')).toEqual({ args: ['hold', undefined, []] });
  });

  it('parses reset', () => {
    expect(parseDeuceswildCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
    expect(parseDeuceswildCommand('reset')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseDeuceswildCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
