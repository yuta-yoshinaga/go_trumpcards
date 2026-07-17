import { describe, expect, it } from 'vitest';
import { parseSpideretteCommand } from './spideretteCommands';

describe('parseSpideretteCommand', () => {
  it('parses deal', () => {
    expect(parseSpideretteCommand('d')).toEqual({ args: ['deal'] });
    expect(parseSpideretteCommand('deal')).toEqual({ args: ['deal'] });
  });

  it('parses move from col to col', () => {
    expect(parseSpideretteCommand('m 0 3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses move with card index', () => {
    expect(parseSpideretteCommand('m 0 2 3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('returns error for move without enough args', () => {
    expect('error' in parseSpideretteCommand('m')).toBe(true);
    expect('error' in parseSpideretteCommand('m 0')).toBe(true);
  });

  it('returns error for move with non-numeric args', () => {
    expect('error' in parseSpideretteCommand('m a b')).toBe(true);
    expect('error' in parseSpideretteCommand('m 0 b 3')).toBe(true);
    expect('error' in parseSpideretteCommand('m 0 2 c')).toBe(true);
  });

  it('parses giveup', () => {
    expect(parseSpideretteCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseSpideretteCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses autocomplete', () => {
    expect(parseSpideretteCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseSpideretteCommand('autocomplete')).toEqual({ args: ['autocomplete'] });
  });

  it('parses undo', () => {
    expect(parseSpideretteCommand('u')).toEqual({ args: ['undo'] });
    expect(parseSpideretteCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses hint', () => {
    expect(parseSpideretteCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSpideretteCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseSpideretteCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseSpideretteCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSpideretteCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for a close typo', () => {
    const result = parseSpideretteCommand('deak');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('deal');
  });

  it('returns error for unknown command', () => {
    expect('error' in parseSpideretteCommand('xyz')).toBe(true);
  });
});
