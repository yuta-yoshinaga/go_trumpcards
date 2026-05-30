import { describe, expect, it } from 'vitest';
import { parseDeuceToSevenCommand } from './deuceToSevenCommands';

describe('parseDeuceToSevenCommand', () => {
  it('parses betting commands', () => {
    expect(parseDeuceToSevenCommand('f')).toEqual({ args: ['fold'] });
    expect(parseDeuceToSevenCommand('fold')).toEqual({ args: ['fold'] });
    expect(parseDeuceToSevenCommand('ck')).toEqual({ args: ['check'] });
    expect(parseDeuceToSevenCommand('c')).toEqual({ args: ['call'] });
    expect(parseDeuceToSevenCommand('a')).toEqual({ args: ['allin'] });
  });

  it('parses bet/raise with amount', () => {
    expect(parseDeuceToSevenCommand('b 50')).toEqual({ args: ['bet', undefined, 50] });
    expect(parseDeuceToSevenCommand('ra 100')).toEqual({ args: ['raise', undefined, 100] });
  });

  it('returns error for bet without amount', () => {
    expect('error' in parseDeuceToSevenCommand('b')).toBe(true);
    expect('error' in parseDeuceToSevenCommand('ra')).toBe(true);
  });

  it('parses exchange with indices', () => {
    expect(parseDeuceToSevenCommand('ex 0 2 4')).toEqual({ args: ['exchange', [0, 2, 4]] });
    expect(parseDeuceToSevenCommand('exchange 1')).toEqual({ args: ['exchange', [1]] });
    expect(parseDeuceToSevenCommand('ex')).toEqual({ args: ['exchange', []] });
  });

  it('returns error for invalid exchange indices', () => {
    expect('error' in parseDeuceToSevenCommand('ex abc')).toBe(true);
  });

  it('parses stand', () => {
    expect(parseDeuceToSevenCommand('st')).toEqual({ args: ['stand'] });
    expect(parseDeuceToSevenCommand('stand')).toEqual({ args: ['stand'] });
  });

  it('parses reset', () => {
    expect(parseDeuceToSevenCommand('r')).toEqual({ args: ['reset'] });
    expect(parseDeuceToSevenCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for mistyped command', () => {
    const result = parseDeuceToSevenCommand('foldd');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseDeuceToSevenCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
