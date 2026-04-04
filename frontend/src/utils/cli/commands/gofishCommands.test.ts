import { describe, expect, it } from 'vitest';
import { parseGofishCommand } from './gofishCommands';

describe('parseGofishCommand', () => {
  it('parses ask with target and rank', () => {
    expect(parseGofishCommand('ask 1 5')).toEqual({ args: ['ask', 1, 5] });
    expect(parseGofishCommand('ask 0 13')).toEqual({ args: ['ask', 0, 13] });
  });

  it('returns error for ask without enough args', () => {
    const result = parseGofishCommand('ask');
    expect('error' in result).toBe(true);
    const result2 = parseGofishCommand('ask 1');
    expect('error' in result2).toBe(true);
  });

  it('parses log', () => {
    expect(parseGofishCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseGofishCommand('r')).toEqual({ args: ['reset'] });
    expect(parseGofishCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseGofishCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
