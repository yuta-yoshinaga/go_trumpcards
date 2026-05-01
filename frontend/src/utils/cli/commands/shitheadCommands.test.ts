import { describe, expect, it } from 'vitest';
import { parseShitheadCommand } from './shitheadCommands';

describe('parseShitheadCommand', () => {
  it('parses play with indices', () => {
    expect(parseShitheadCommand('p 0 2 3')).toEqual({
      args: ['play', { indices: [0, 2, 3] }],
    });
    expect(parseShitheadCommand('play 1')).toEqual({
      args: ['play', { indices: [1] }],
    });
  });

  it('parses pickup', () => {
    expect(parseShitheadCommand('pu')).toEqual({ args: ['play', { indices: [] }] });
    expect(parseShitheadCommand('pickup')).toEqual({ args: ['play', { indices: [] }] });
  });

  it('rejects play with non-numeric args', () => {
    expect('error' in parseShitheadCommand('p abc')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseShitheadCommand('log')).toEqual({ args: ['log'] });
    expect(parseShitheadCommand('r')).toEqual({ args: ['reset'] });
    expect(parseShitheadCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseShitheadCommand('pickupp');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseShitheadCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
