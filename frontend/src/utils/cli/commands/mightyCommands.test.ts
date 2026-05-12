import { describe, expect, it } from 'vitest';
import { parseMightyCommand } from './mightyCommands';

describe('parseMightyCommand', () => {
  it('parses bid with number', () => {
    expect(parseMightyCommand('bid 14')).toEqual({ args: ['bid', 14, false] });
  });

  it('parses bid with No-Trump modifier', () => {
    expect(parseMightyCommand('bid 14 nt')).toEqual({ args: ['bid', 14, true] });
  });

  it('returns error for bid without number', () => {
    const result = parseMightyCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('parses pass', () => {
    expect(parseMightyCommand('pass')).toEqual({ args: ['bid', 0, false] });
  });

  it('parses trump with three args', () => {
    expect(parseMightyCommand('trump spade heart 1')).toEqual({
      args: ['trump', undefined, undefined, undefined, 1, 3, 1],
    });
  });

  it('parses trump with no-trump declaration', () => {
    expect(parseMightyCommand('trump nt heart 13')).toEqual({
      args: ['trump', undefined, undefined, undefined, -1, 3, 13],
    });
  });

  it('parses trump with joker partner', () => {
    expect(parseMightyCommand('trump spade joker 0')).toEqual({
      args: ['trump', undefined, undefined, undefined, 1, 0, 0],
    });
  });

  it('returns error for trump without enough args', () => {
    const result = parseMightyCommand('trump spade');
    expect('error' in result).toBe(true);
  });

  it('returns error for trump with invalid suit', () => {
    const result = parseMightyCommand('trump bogus heart 1');
    expect('error' in result).toBe(true);
  });

  it('parses exchange with three indices', () => {
    expect(parseMightyCommand('ex 0 1 2')).toEqual({
      args: ['exchange', undefined, undefined, undefined, undefined, undefined, undefined, [0, 1, 2]],
    });
  });

  it('parses exchange long form', () => {
    expect(parseMightyCommand('exchange 3 4 5')).toEqual({
      args: ['exchange', undefined, undefined, undefined, undefined, undefined, undefined, [3, 4, 5]],
    });
  });

  it('returns error for exchange with too few indices', () => {
    const result = parseMightyCommand('ex 0 1');
    expect('error' in result).toBe(true);
  });

  it('parses jokerlead', () => {
    expect(parseMightyCommand('jl 0 heart')).toEqual({
      args: ['jokerlead', undefined, undefined, 0, undefined, undefined, undefined, undefined, 3],
    });
  });

  it('returns error for jokerlead with joker suit', () => {
    const result = parseMightyCommand('jl 0 joker');
    expect('error' in result).toBe(true);
  });

  it('returns error for jokerlead with too few args', () => {
    const result = parseMightyCommand('jl');
    expect('error' in result).toBe(true);
  });

  it('parses log', () => {
    expect(parseMightyCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses play from shared trick commands', () => {
    expect(parseMightyCommand('p 3')).toEqual({ args: ['play', undefined, undefined, 3] });
  });

  it('parses reset', () => {
    expect(parseMightyCommand('r')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseMightyCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
