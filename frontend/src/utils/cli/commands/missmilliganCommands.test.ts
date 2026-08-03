import { describe, expect, it } from 'vitest';
import { parseMissMilliganCommand } from './missmilliganCommands';

describe('parseMissMilliganCommand', () => {
  it('parses deal', () => {
    expect(parseMissMilliganCommand('d')).toEqual({ args: ['deal'] });
    expect(parseMissMilliganCommand('deal')).toEqual({ args: ['deal'] });
  });

  it('parses waive with and without a run head', () => {
    expect(parseMissMilliganCommand('wv t3')).toEqual({
      args: ['waive', { zone: 'tableau', col: 3, cardIndex: undefined }],
    });
    expect(parseMissMilliganCommand('waive t3.1')).toEqual({
      args: ['waive', { zone: 'tableau', col: 3, cardIndex: 1 }],
    });
  });

  it('parses tableau moves', () => {
    expect(parseMissMilliganCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
    expect(parseMissMilliganCommand('m t0 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'tableau', col: 5 }],
    });
    expect(parseMissMilliganCommand('m t0.2 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('parses waived moves', () => {
    expect(parseMissMilliganCommand('m w f')).toEqual({
      args: ['move', { zone: 'waived' }, { zone: 'foundation' }],
    });
    expect(parseMissMilliganCommand('m w t3')).toEqual({
      args: ['move', { zone: 'waived' }, { zone: 'tableau', col: 3 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseMissMilliganCommand('m')).toBe(true);
    expect('error' in parseMissMilliganCommand('m t0')).toBe(true);
    expect('error' in parseMissMilliganCommand('wv')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseMissMilliganCommand('m x0 t1')).toBe(true);
    expect('error' in parseMissMilliganCommand('m tz t1')).toBe(true);
    expect('error' in parseMissMilliganCommand('m t t1')).toBe(true);
    expect('error' in parseMissMilliganCommand('m t0.z t1')).toBe(true);
    expect('error' in parseMissMilliganCommand('m t0 z')).toBe(true);
    expect('error' in parseMissMilliganCommand('m t0 tz')).toBe(true);
    expect('error' in parseMissMilliganCommand('m w z')).toBe(true);
    expect('error' in parseMissMilliganCommand('m w tz')).toBe(true);
    expect('error' in parseMissMilliganCommand('wv x3')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseMissMilliganCommand('u')).toEqual({ args: ['undo'] });
    expect(parseMissMilliganCommand('h')).toEqual({ args: ['hint'] });
    expect(parseMissMilliganCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseMissMilliganCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseMissMilliganCommand('log')).toEqual({ args: ['log'] });
    expect(parseMissMilliganCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseMissMilliganCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseMissMilliganCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
