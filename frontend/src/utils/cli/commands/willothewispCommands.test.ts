import { describe, expect, it } from 'vitest';
import { parseWillOTheWispCommand } from './willothewispCommands';

describe('parseWillOTheWispCommand', () => {
  it('parses deal', () => {
    expect(parseWillOTheWispCommand('d')).toEqual({ args: ['deal'] });
    expect(parseWillOTheWispCommand('deal')).toEqual({ args: ['deal'] });
  });

  it('parses move from col to col', () => {
    expect(parseWillOTheWispCommand('m 0 3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses move with card index', () => {
    expect(parseWillOTheWispCommand('m 0 2 3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('returns error for move without enough args', () => {
    expect('error' in parseWillOTheWispCommand('m')).toBe(true);
    expect('error' in parseWillOTheWispCommand('m 0')).toBe(true);
  });

  it('returns error for move with non-numeric args', () => {
    expect('error' in parseWillOTheWispCommand('m a b')).toBe(true);
    expect('error' in parseWillOTheWispCommand('m 0 b 3')).toBe(true);
    expect('error' in parseWillOTheWispCommand('m 0 2 c')).toBe(true);
  });

  it('parses giveup', () => {
    expect(parseWillOTheWispCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseWillOTheWispCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses autocomplete', () => {
    expect(parseWillOTheWispCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseWillOTheWispCommand('autocomplete')).toEqual({ args: ['autocomplete'] });
  });

  it('parses undo', () => {
    expect(parseWillOTheWispCommand('u')).toEqual({ args: ['undo'] });
    expect(parseWillOTheWispCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses hint', () => {
    expect(parseWillOTheWispCommand('h')).toEqual({ args: ['hint'] });
    expect(parseWillOTheWispCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseWillOTheWispCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseWillOTheWispCommand('r')).toEqual({ args: ['reset'] });
    expect(parseWillOTheWispCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for a close typo', () => {
    const result = parseWillOTheWispCommand('deak');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('deal');
  });

  it('returns error for unknown command', () => {
    expect('error' in parseWillOTheWispCommand('xyz')).toBe(true);
  });
});
