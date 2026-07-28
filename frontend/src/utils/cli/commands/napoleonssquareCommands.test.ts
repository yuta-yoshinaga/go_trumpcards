import { describe, expect, it } from 'vitest';
import { parseNapoleonsSquareCommand } from './napoleonssquareCommands';

describe('parseNapoleonsSquareCommand', () => {
  it('parses waste moves', () => {
    expect(parseNapoleonsSquareCommand('m w f')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'foundation' }],
    });
    expect(parseNapoleonsSquareCommand('m w t3')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses tableau to foundation', () => {
    expect(parseNapoleonsSquareCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
  });

  it('moves only the top card when no run head is given', () => {
    expect(parseNapoleonsSquareCommand('m t0 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'tableau', col: 5 }],
    });
  });

  it('carries a run when the head index is given', () => {
    expect(parseNapoleonsSquareCommand('m t0.2 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseNapoleonsSquareCommand('m')).toBe(true);
    expect('error' in parseNapoleonsSquareCommand('m t0')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseNapoleonsSquareCommand('m x0 t1')).toBe(true);
    expect('error' in parseNapoleonsSquareCommand('m tz t1')).toBe(true);
    expect('error' in parseNapoleonsSquareCommand('m t0.z t1')).toBe(true);
    expect('error' in parseNapoleonsSquareCommand('m t0 z')).toBe(true);
    expect('error' in parseNapoleonsSquareCommand('m t0 tz')).toBe(true);
    expect('error' in parseNapoleonsSquareCommand('m w z')).toBe(true);
    expect('error' in parseNapoleonsSquareCommand('m w tz')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseNapoleonsSquareCommand('d')).toEqual({ args: ['draw'] });
    expect(parseNapoleonsSquareCommand('draw')).toEqual({ args: ['draw'] });
    expect(parseNapoleonsSquareCommand('u')).toEqual({ args: ['undo'] });
    expect(parseNapoleonsSquareCommand('h')).toEqual({ args: ['hint'] });
    expect(parseNapoleonsSquareCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseNapoleonsSquareCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseNapoleonsSquareCommand('log')).toEqual({ args: ['log'] });
    expect(parseNapoleonsSquareCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseNapoleonsSquareCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseNapoleonsSquareCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
