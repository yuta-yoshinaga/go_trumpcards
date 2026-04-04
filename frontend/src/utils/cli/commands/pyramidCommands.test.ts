import { describe, expect, it } from 'vitest';
import { parsePyramidCommand } from './pyramidCommands';

describe('parsePyramidCommand', () => {
  it('parses draw', () => {
    expect(parsePyramidCommand('d')).toEqual({ args: ['draw'] });
    expect(parsePyramidCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses remove pair of pyramid cards', () => {
    expect(parsePyramidCommand('rm 0 0 1 1')).toEqual({
      args: ['remove', { zone: 'pyramid', row: 0, col: 0 }, { zone: 'pyramid', row: 1, col: 1 }],
    });
  });

  it('parses remove pyramid card with waste', () => {
    expect(parsePyramidCommand('rm 2 3 w')).toEqual({
      args: ['remove', { zone: 'pyramid', row: 2, col: 3 }, { zone: 'waste' }],
    });
  });

  it('returns error for remove without enough args', () => {
    const result = parsePyramidCommand('rm');
    expect('error' in result).toBe(true);
    const result2 = parsePyramidCommand('rm 0');
    expect('error' in result2).toBe(true);
  });

  it('returns error for remove with only 3 non-waste args', () => {
    const result = parsePyramidCommand('rm 0 0 1');
    expect('error' in result).toBe(true);
  });

  it('parses king from pyramid', () => {
    expect(parsePyramidCommand('k 3 2')).toEqual({
      args: ['remove', { zone: 'pyramid', row: 3, col: 2 }],
    });
    expect(parsePyramidCommand('king 5 0')).toEqual({
      args: ['remove', { zone: 'pyramid', row: 5, col: 0 }],
    });
  });

  it('parses king from waste', () => {
    expect(parsePyramidCommand('k w 0')).toEqual({
      args: ['remove', { zone: 'waste' }],
    });
  });

  it('returns error for king without enough args', () => {
    const result = parsePyramidCommand('k');
    expect('error' in result).toBe(true);
    const result2 = parsePyramidCommand('k 3');
    expect('error' in result2).toBe(true);
  });

  it('parses giveup', () => {
    expect(parsePyramidCommand('g')).toEqual({ args: ['giveup'] });
    expect(parsePyramidCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses hint', () => {
    expect(parsePyramidCommand('h')).toEqual({ args: ['hint'] });
    expect(parsePyramidCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parsePyramidCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses undo', () => {
    expect(parsePyramidCommand('u')).toEqual({ args: ['undo'] });
    expect(parsePyramidCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses reset', () => {
    expect(parsePyramidCommand('r')).toEqual({ args: ['reset'] });
    expect(parsePyramidCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parsePyramidCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
