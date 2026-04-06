import { describe, expect, it } from 'vitest';
import { parseVideopokerCommand } from './videopokerCommands';

describe('parseVideopokerCommand', () => {
  it('parses bet with amount', () => {
    expect(parseVideopokerCommand('b 100')).toEqual({ args: ['bet', 100, undefined] });
    expect(parseVideopokerCommand('bet 50')).toEqual({ args: ['bet', 50, undefined] });
  });

  it('returns error for bet without amount', () => {
    const result = parseVideopokerCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses hold with indices', () => {
    expect(parseVideopokerCommand('hold 0 2 4')).toEqual({ args: ['hold', undefined, [0, 2, 4]] });
  });

  it('parses hold without indices as empty', () => {
    expect(parseVideopokerCommand('hold')).toEqual({ args: ['hold', undefined, []] });
  });

  it('parses reset', () => {
    expect(parseVideopokerCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
    expect(parseVideopokerCommand('reset')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseVideopokerCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
