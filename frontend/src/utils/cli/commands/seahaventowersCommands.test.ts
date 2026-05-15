import { describe, expect, it } from 'vitest';
import { parseSeahavenTowersCommand } from './seahaventowersCommands';

describe('parseSeahavenTowersCommand', () => {
  it('parses move tableau to foundation', () => {
    expect(parseSeahavenTowersCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
  });

  it('parses move tableau to tableau', () => {
    expect(parseSeahavenTowersCommand('m t0 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses move tableau to reserved cell', () => {
    expect(parseSeahavenTowersCommand('m t0 c1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'reserved', cell: 1 }],
    });
  });

  it('parses move tableau with card index to tableau (3 packed args)', () => {
    // The packed form `m t<col> <cardIdx> t<col>` yields exactly 3 args after
    // splitCommand strips the leading `m`; this branch must fire on length >= 3.
    expect(parseSeahavenTowersCommand('m t0 2 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses move reserved cell to foundation', () => {
    expect(parseSeahavenTowersCommand('m c0 f')).toEqual({
      args: ['move', { zone: 'reserved', cell: 0 }, { zone: 'foundation' }],
    });
  });

  it('parses move reserved cell to tableau', () => {
    expect(parseSeahavenTowersCommand('m c1 t2')).toEqual({
      args: ['move', { zone: 'reserved', cell: 1 }, { zone: 'tableau', col: 2 }],
    });
  });

  it('returns error for move without enough args', () => {
    expect('error' in parseSeahavenTowersCommand('m')).toBe(true);
  });

  it('returns error for invalid source zone letter', () => {
    expect('error' in parseSeahavenTowersCommand('m x0 t1')).toBe(true);
  });

  it('returns error for invalid tableau column number', () => {
    expect('error' in parseSeahavenTowersCommand('m tx t1')).toBe(true);
  });

  it('returns error for invalid reserved cell number', () => {
    expect('error' in parseSeahavenTowersCommand('m cx f')).toBe(true);
  });

  it('returns error for invalid target from tableau', () => {
    expect('error' in parseSeahavenTowersCommand('m t0 x')).toBe(true);
  });

  it('returns error for invalid target from reserved cell', () => {
    expect('error' in parseSeahavenTowersCommand('m c0 x')).toBe(true);
  });

  it('returns error for invalid target tableau column', () => {
    expect('error' in parseSeahavenTowersCommand('m t0 tx')).toBe(true);
  });

  it('returns error for invalid target cell index', () => {
    expect('error' in parseSeahavenTowersCommand('m t0 cx')).toBe(true);
  });

  it('returns error for invalid card index in packed form', () => {
    expect('error' in parseSeahavenTowersCommand('m t0 abc t3')).toBe(true);
  });

  it('returns error for invalid target col in packed form', () => {
    expect('error' in parseSeahavenTowersCommand('m t0 2 tx')).toBe(true);
  });

  it('returns error for invalid target col from reserved', () => {
    expect('error' in parseSeahavenTowersCommand('m c0 tx')).toBe(true);
  });

  it('parses giveup', () => {
    expect(parseSeahavenTowersCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseSeahavenTowersCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses autocomplete', () => {
    expect(parseSeahavenTowersCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseSeahavenTowersCommand('autocomplete')).toEqual({ args: ['autocomplete'] });
  });

  it('parses undo', () => {
    expect(parseSeahavenTowersCommand('u')).toEqual({ args: ['undo'] });
    expect(parseSeahavenTowersCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses hint', () => {
    expect(parseSeahavenTowersCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSeahavenTowersCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseSeahavenTowersCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseSeahavenTowersCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSeahavenTowersCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a near match for typo', () => {
    const result = parseSeahavenTowersCommand('mvoe t0 f');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toMatch(/did you mean/i);
  });

  it('returns plain error for far-off unknown command', () => {
    const result = parseSeahavenTowersCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
