import { describe, expect, it } from 'vitest';
import { parseIndianpokerCommand } from './indianpokerCommands';

describe('parseIndianpokerCommand', () => {
  it('parses fold', () => {
    expect(parseIndianpokerCommand('f')).toEqual({ args: ['fold', undefined] });
    expect(parseIndianpokerCommand('fold')).toEqual({ args: ['fold', undefined] });
  });

  it('parses check', () => {
    expect(parseIndianpokerCommand('ck')).toEqual({ args: ['check', undefined] });
  });

  it('parses call', () => {
    expect(parseIndianpokerCommand('c')).toEqual({ args: ['call', undefined] });
  });

  it('parses bet with amount', () => {
    expect(parseIndianpokerCommand('b 100')).toEqual({ args: ['bet', 100] });
  });

  it('returns error for bet without amount', () => {
    const result = parseIndianpokerCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses raise with amount', () => {
    expect(parseIndianpokerCommand('ra 200')).toEqual({ args: ['raise', 200] });
  });

  it('parses allin', () => {
    expect(parseIndianpokerCommand('a')).toEqual({ args: ['allin', undefined] });
  });

  it('parses log', () => {
    expect(parseIndianpokerCommand('log')).toEqual({ args: ['log', undefined] });
  });

  it('parses reset', () => {
    expect(parseIndianpokerCommand('r')).toEqual({ args: ['reset', undefined] });
    expect(parseIndianpokerCommand('reset')).toEqual({ args: ['reset', undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseIndianpokerCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
