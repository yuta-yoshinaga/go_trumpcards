import { describe, expect, it } from 'vitest';
import { parseStalactitesCommand } from './stalactitesCommands';

describe('parseStalactitesCommand', () => {
  it('parses move tableau to foundation', () => {
    expect(parseStalactitesCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
  });

  it('parses move tableau to tableau', () => {
    expect(parseStalactitesCommand('m t0 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses move tableau to cell', () => {
    expect(parseStalactitesCommand('m t0 c1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'cell', cell: 1 }],
    });
  });

  it('parses move tableau with card index to tableau (4 args)', () => {
    // Code requires args.length >= 4 for cardIndex path
    expect(parseStalactitesCommand('m t0 2 t3 0')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses move cell to foundation', () => {
    expect(parseStalactitesCommand('m c0 f')).toEqual({
      args: ['move', { zone: 'cell', cell: 0 }, { zone: 'foundation' }],
    });
  });

  it('parses move cell to tableau', () => {
    expect(parseStalactitesCommand('m c0 t2')).toEqual({
      args: ['move', { zone: 'cell', cell: 0 }, { zone: 'tableau', col: 2 }],
    });
  });

  it('returns error for move without enough args', () => {
    const result = parseStalactitesCommand('m');
    expect('error' in result).toBe(true);
  });

  it('returns error for invalid source zone', () => {
    const result = parseStalactitesCommand('m x0 t1');
    expect('error' in result).toBe(true);
  });

  it('returns error for invalid target from tableau', () => {
    const result = parseStalactitesCommand('m t0 x');
    expect('error' in result).toBe(true);
  });

  it('returns error for invalid target from cell', () => {
    const result = parseStalactitesCommand('m c0 x');
    expect('error' in result).toBe(true);
  });

  it('parses giveup', () => {
    expect(parseStalactitesCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseStalactitesCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses autocomplete', () => {
    expect(parseStalactitesCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseStalactitesCommand('autocomplete')).toEqual({ args: ['autocomplete'] });
  });

  it('parses undo', () => {
    expect(parseStalactitesCommand('u')).toEqual({ args: ['undo'] });
    expect(parseStalactitesCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses hint', () => {
    expect(parseStalactitesCommand('h')).toEqual({ args: ['hint'] });
    expect(parseStalactitesCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseStalactitesCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseStalactitesCommand('r')).toEqual({ args: ['reset'] });
    expect(parseStalactitesCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseStalactitesCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
