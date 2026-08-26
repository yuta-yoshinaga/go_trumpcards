import { describe, expect, it } from 'vitest';
import { parseMrsMopCommand } from './mrsMopCommands';

describe('parseMrsMopCommand', () => {
  // **配るコマンドは無い。**104枚を配り切るので山札が存在しない。
  // クローン元 (Spider) は `d`/`deal` で5回配れる。
  it('rejects deal, which this game has no stock for', () => {
    expect('error' in parseMrsMopCommand('d')).toBe(true);
    expect('error' in parseMrsMopCommand('deal')).toBe(true);
  });

  it('parses move from col to col', () => {
    expect(parseMrsMopCommand('m 0 3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses move with card index', () => {
    expect(parseMrsMopCommand('m 0 2 3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('returns error for move without enough args', () => {
    const result = parseMrsMopCommand('m');
    expect('error' in result).toBe(true);
    const result2 = parseMrsMopCommand('m 0');
    expect('error' in result2).toBe(true);
  });

  it('parses giveup', () => {
    expect(parseMrsMopCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseMrsMopCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses autocomplete', () => {
    expect(parseMrsMopCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseMrsMopCommand('autocomplete')).toEqual({ args: ['autocomplete'] });
  });

  it('parses undo', () => {
    expect(parseMrsMopCommand('u')).toEqual({ args: ['undo'] });
    expect(parseMrsMopCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses hint', () => {
    expect(parseMrsMopCommand('h')).toEqual({ args: ['hint'] });
    expect(parseMrsMopCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseMrsMopCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset without args', () => {
    expect(parseMrsMopCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMrsMopCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('parses reset with difficulty', () => {
    expect(parseMrsMopCommand('r 1')).toEqual({ args: ['reset', undefined, undefined, { difficulty: 1 }] });
    expect(parseMrsMopCommand('r 4')).toEqual({ args: ['reset', undefined, undefined, { difficulty: 4 }] });
  });

  it('returns error for unknown command', () => {
    const result = parseMrsMopCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
