import { describe, expect, it } from 'vitest';
import { parseJokerpokerCommand } from './jokerpokerCommands';

describe('parseJokerpokerCommand', () => {
  it('parses bet with amount', () => {
    expect(parseJokerpokerCommand('b 100')).toEqual({ args: ['bet', 100, undefined] });
    expect(parseJokerpokerCommand('bet 50')).toEqual({ args: ['bet', 50, undefined] });
  });

  it('returns error for bet without amount', () => {
    const result = parseJokerpokerCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses hold with indices', () => {
    expect(parseJokerpokerCommand('hold 0 2 4')).toEqual({ args: ['hold', undefined, [0, 2, 4]] });
  });

  it('parses hold without indices as empty', () => {
    expect(parseJokerpokerCommand('hold')).toEqual({ args: ['hold', undefined, []] });
  });

  it('parses reset', () => {
    expect(parseJokerpokerCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
    expect(parseJokerpokerCommand('reset')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseJokerpokerCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
