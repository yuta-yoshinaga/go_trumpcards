import { describe, expect, it } from 'vitest';
import { parseTerraceCommand } from './terraceCommands';

describe('parseTerraceCommand', () => {
  it('parses draw', () => {
    expect(parseTerraceCommand('d')).toEqual({ args: ['draw'] });
    expect(parseTerraceCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses the terrace to a foundation', () => {
    expect(parseTerraceCommand('m r f')).toEqual({
      args: ['move', { zone: 'reserve' }, { zone: 'foundation' }],
    });
  });

  // The terrace has exactly one destination.
  it('rejects the terrace going to a pile', () => {
    const r = parseTerraceCommand('m r t2');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('foundation');
  });

  it('parses waste moves', () => {
    expect(parseTerraceCommand('m w f')).toEqual({ args: ['move', { zone: 'waste' }, { zone: 'foundation' }] });
    expect(parseTerraceCommand('m w t2')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 2 }],
    });
  });

  it('parses tableau moves', () => {
    expect(parseTerraceCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
    expect(parseTerraceCommand('m t0 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseTerraceCommand('m')).toBe(true);
    expect('error' in parseTerraceCommand('m t0')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseTerraceCommand('m x0 f')).toBe(true);
    expect('error' in parseTerraceCommand('m tz f')).toBe(true);
    expect('error' in parseTerraceCommand('m t f')).toBe(true);
    expect('error' in parseTerraceCommand('m t0 z')).toBe(true);
    expect('error' in parseTerraceCommand('m t0 tz')).toBe(true);
    expect('error' in parseTerraceCommand('m t0 t')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseTerraceCommand('u')).toEqual({ args: ['undo'] });
    expect(parseTerraceCommand('h')).toEqual({ args: ['hint'] });
    expect(parseTerraceCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseTerraceCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseTerraceCommand('log')).toEqual({ args: ['log'] });
    expect(parseTerraceCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseTerraceCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseTerraceCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
