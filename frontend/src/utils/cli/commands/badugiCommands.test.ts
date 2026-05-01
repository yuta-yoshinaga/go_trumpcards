import { describe, expect, it } from 'vitest';
import { parseBadugiCommand } from './badugiCommands';

describe('parseBadugiCommand', () => {
  it('parses betting commands', () => {
    expect(parseBadugiCommand('f')).toEqual({ args: ['fold'] });
    expect(parseBadugiCommand('fold')).toEqual({ args: ['fold'] });
    expect(parseBadugiCommand('ck')).toEqual({ args: ['check'] });
    expect(parseBadugiCommand('c')).toEqual({ args: ['call'] });
    expect(parseBadugiCommand('a')).toEqual({ args: ['allin'] });
  });

  it('parses bet/raise with amount', () => {
    expect(parseBadugiCommand('b 50')).toEqual({ args: ['bet', undefined, 50] });
    expect(parseBadugiCommand('ra 100')).toEqual({ args: ['raise', undefined, 100] });
  });

  it('returns error for bet without amount', () => {
    expect('error' in parseBadugiCommand('b')).toBe(true);
    expect('error' in parseBadugiCommand('ra')).toBe(true);
  });

  it('parses exchange with indices', () => {
    expect(parseBadugiCommand('ex 0 2 3')).toEqual({ args: ['exchange', [0, 2, 3]] });
    expect(parseBadugiCommand('exchange 1')).toEqual({ args: ['exchange', [1]] });
    expect(parseBadugiCommand('ex')).toEqual({ args: ['exchange', []] });
  });

  it('returns error for invalid exchange indices', () => {
    expect('error' in parseBadugiCommand('ex abc')).toBe(true);
  });

  it('parses stand', () => {
    expect(parseBadugiCommand('st')).toEqual({ args: ['stand'] });
    expect(parseBadugiCommand('stand')).toEqual({ args: ['stand'] });
  });

  it('parses reset', () => {
    expect(parseBadugiCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBadugiCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for mistyped command', () => {
    const result = parseBadugiCommand('foldd');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseBadugiCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
