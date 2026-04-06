import { describe, expect, it } from 'vitest';
import { parseSpiderCommand } from './spiderCommands';

describe('parseSpiderCommand', () => {
  it('parses deal', () => {
    expect(parseSpiderCommand('d')).toEqual({ args: ['deal'] });
    expect(parseSpiderCommand('deal')).toEqual({ args: ['deal'] });
  });

  it('parses move from col to col', () => {
    expect(parseSpiderCommand('m 0 3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses move with card index', () => {
    expect(parseSpiderCommand('m 0 2 3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('returns error for move without enough args', () => {
    const result = parseSpiderCommand('m');
    expect('error' in result).toBe(true);
    const result2 = parseSpiderCommand('m 0');
    expect('error' in result2).toBe(true);
  });

  it('parses giveup', () => {
    expect(parseSpiderCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseSpiderCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses autocomplete', () => {
    expect(parseSpiderCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseSpiderCommand('autocomplete')).toEqual({ args: ['autocomplete'] });
  });

  it('parses undo', () => {
    expect(parseSpiderCommand('u')).toEqual({ args: ['undo'] });
    expect(parseSpiderCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses hint', () => {
    expect(parseSpiderCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSpiderCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseSpiderCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset without args', () => {
    expect(parseSpiderCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSpiderCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('parses reset with difficulty', () => {
    expect(parseSpiderCommand('r 1')).toEqual({ args: ['reset', undefined, undefined, { difficulty: 1 }] });
    expect(parseSpiderCommand('r 4')).toEqual({ args: ['reset', undefined, undefined, { difficulty: 4 }] });
  });

  it('returns error for unknown command', () => {
    const result = parseSpiderCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
