import { describe, expect, it } from 'vitest';
import { parseDaifugoCommand } from './daifugoCommands';

describe('parseDaifugoCommand', () => {
  it('parses play with indices', () => {
    expect(parseDaifugoCommand('p 0 2')).toEqual({ args: ['play', [0, 2]] });
  });

  it('parses pass (no args)', () => {
    expect(parseDaifugoCommand('p')).toEqual({ args: ['play'] });
  });

  it('parses sort', () => {
    expect(parseDaifugoCommand('sort')).toEqual({ args: ['sort', undefined, undefined, 0] });
    expect(parseDaifugoCommand('sort 1')).toEqual({ args: ['sort', undefined, undefined, 1] });
  });

  it('parses reset', () => {
    expect(parseDaifugoCommand('r')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseDaifugoCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
