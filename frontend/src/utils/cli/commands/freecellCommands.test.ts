import { describe, expect, it } from 'vitest';
import { parseFreecellCommand } from './freecellCommands';

describe('parseFreecellCommand', () => {
  it('parses move tableau to foundation', () => {
    expect(parseFreecellCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
  });

  it('parses move tableau to tableau', () => {
    expect(parseFreecellCommand('m t0 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses move tableau to cell', () => {
    expect(parseFreecellCommand('m t0 c1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'cell', cell: 1 }],
    });
  });

  it('parses move tableau with card index to tableau (4 args)', () => {
    // Code requires args.length >= 4 for cardIndex path
    expect(parseFreecellCommand('m t0 2 t3 0')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses move cell to foundation', () => {
    expect(parseFreecellCommand('m c0 f')).toEqual({
      args: ['move', { zone: 'cell', cell: 0 }, { zone: 'foundation' }],
    });
  });

  it('parses move cell to tableau', () => {
    expect(parseFreecellCommand('m c0 t2')).toEqual({
      args: ['move', { zone: 'cell', cell: 0 }, { zone: 'tableau', col: 2 }],
    });
  });

  it('returns error for move without enough args', () => {
    const result = parseFreecellCommand('m');
    expect('error' in result).toBe(true);
  });

  it('returns error for invalid source zone', () => {
    const result = parseFreecellCommand('m x0 t1');
    expect('error' in result).toBe(true);
  });

  it('returns error for invalid target from tableau', () => {
    const result = parseFreecellCommand('m t0 x');
    expect('error' in result).toBe(true);
  });

  it('returns error for invalid target from cell', () => {
    const result = parseFreecellCommand('m c0 x');
    expect('error' in result).toBe(true);
  });

  it('parses giveup', () => {
    expect(parseFreecellCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseFreecellCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses autocomplete', () => {
    expect(parseFreecellCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseFreecellCommand('autocomplete')).toEqual({ args: ['autocomplete'] });
  });

  it('parses undo', () => {
    expect(parseFreecellCommand('u')).toEqual({ args: ['undo'] });
    expect(parseFreecellCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses hint', () => {
    expect(parseFreecellCommand('h')).toEqual({ args: ['hint'] });
    expect(parseFreecellCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseFreecellCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseFreecellCommand('r')).toEqual({ args: ['reset'] });
    expect(parseFreecellCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseFreecellCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
