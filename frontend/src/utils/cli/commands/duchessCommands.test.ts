import { describe, expect, it } from 'vitest';
import { parseDuchessCommand } from './duchessCommands';

describe('parseDuchessCommand', () => {
  it('parses draw', () => {
    expect(parseDuchessCommand('d')).toEqual({ args: ['draw'] });
    expect(parseDuchessCommand('draw')).toEqual({ args: ['draw'] });
  });

  // Choosing the base rank is its own command, not a move.
  it('parses the base-rank command', () => {
    expect(parseDuchessCommand('b 2')).toEqual({ args: ['base', { zone: 'reserve', col: 2 }] });
    expect(parseDuchessCommand('base 0')).toEqual({ args: ['base', { zone: 'reserve', col: 0 }] });
  });

  it('parses reserve moves', () => {
    expect(parseDuchessCommand('m r1 f')).toEqual({
      args: ['move', { zone: 'reserve', col: 1 }, { zone: 'foundation' }],
    });
    expect(parseDuchessCommand('m r3 t0')).toEqual({
      args: ['move', { zone: 'reserve', col: 3 }, { zone: 'tableau', col: 0 }],
    });
  });

  it('parses waste moves', () => {
    expect(parseDuchessCommand('m w f')).toEqual({ args: ['move', { zone: 'waste' }, { zone: 'foundation' }] });
    expect(parseDuchessCommand('m w t2')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 2 }],
    });
  });

  it('parses tableau moves', () => {
    expect(parseDuchessCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation' }],
    });
    expect(parseDuchessCommand('m t0 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'tableau', col: 3 }],
    });
    expect(parseDuchessCommand('m t0.2 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseDuchessCommand('m')).toBe(true);
    expect('error' in parseDuchessCommand('m t0')).toBe(true);
    expect('error' in parseDuchessCommand('b')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseDuchessCommand('b x')).toBe(true);
    expect('error' in parseDuchessCommand('m x0 t1')).toBe(true);
    expect('error' in parseDuchessCommand('m rz t1')).toBe(true);
    expect('error' in parseDuchessCommand('m r t1')).toBe(true);
    expect('error' in parseDuchessCommand('m tz t1')).toBe(true);
    expect('error' in parseDuchessCommand('m t t1')).toBe(true);
    expect('error' in parseDuchessCommand('m t0.z t1')).toBe(true);
    expect('error' in parseDuchessCommand('m t0 z')).toBe(true);
    expect('error' in parseDuchessCommand('m t0 tz')).toBe(true);
    expect('error' in parseDuchessCommand('m t0 t')).toBe(true);
    expect('error' in parseDuchessCommand('m w z')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseDuchessCommand('u')).toEqual({ args: ['undo'] });
    expect(parseDuchessCommand('h')).toEqual({ args: ['hint'] });
    expect(parseDuchessCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseDuchessCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseDuchessCommand('log')).toEqual({ args: ['log'] });
    expect(parseDuchessCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseDuchessCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseDuchessCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
