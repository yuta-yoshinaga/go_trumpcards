import { describe, expect, it } from 'vitest';
import { parseNinetynineCommand } from './ninetynineCommands';

describe('parseNinetynineCommand', () => {
  it('parses bid with three indices', () => {
    expect(parseNinetynineCommand('bid 0 1 2')).toEqual({ args: ['bid', [0, 1, 2], undefined] });
  });

  it('returns error for bid without three indices', () => {
    expect('error' in parseNinetynineCommand('bid 0 1')).toBe(true);
    expect('error' in parseNinetynineCommand('bid')).toBe(true);
  });

  it('returns error for bid with non-numeric index', () => {
    expect('error' in parseNinetynineCommand('bid 0 x 2')).toBe(true);
  });

  it('parses play with index', () => {
    expect(parseNinetynineCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
    expect(parseNinetynineCommand('play 5')).toEqual({ args: ['play', undefined, 5] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseNinetynineCommand('p')).toBe(true);
  });

  it('parses next', () => {
    expect(parseNinetynineCommand('n')).toEqual({ args: ['next', undefined, undefined] });
  });

  it('parses nextround', () => {
    expect(parseNinetynineCommand('nr')).toEqual({ args: ['nextround', undefined, undefined] });
  });

  it('parses hint', () => {
    expect(parseNinetynineCommand('h')).toEqual({ args: ['hint', undefined, undefined] });
  });

  it('parses reset', () => {
    expect(parseNinetynineCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parseNinetynineCommand('xyz')).toBe(true);
  });
});
