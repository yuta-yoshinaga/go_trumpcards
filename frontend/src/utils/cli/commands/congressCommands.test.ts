import { describe, expect, it } from 'vitest';
import { parseCongressCommand } from './congressCommands';

describe('parseCongressCommand', () => {
  it('parses draw', () => {
    expect(parseCongressCommand('d')).toEqual({ args: ['draw'] });
    expect(parseCongressCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses tableau moves', () => {
    expect(parseCongressCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
    expect(parseCongressCommand('m t0 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('parses waste moves', () => {
    expect(parseCongressCommand('m w f')).toEqual({ args: ['move', { zone: 'waste' }, { zone: 'foundation' }] });
    expect(parseCongressCommand('m w t2')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 2 }],
    });
  });

  it('parses the stock filling a gap', () => {
    expect(parseCongressCommand('m s t3')).toEqual({
      args: ['move', { zone: 'stock' }, { zone: 'tableau', col: 3 }],
    });
  });

  // The stock never reaches a foundation directly.
  it('rejects the stock going to a foundation', () => {
    const r = parseCongressCommand('m s f');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('empty pile');
  });

  it('returns error for missing args', () => {
    expect('error' in parseCongressCommand('m')).toBe(true);
    expect('error' in parseCongressCommand('m t0')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseCongressCommand('m x0 f')).toBe(true);
    expect('error' in parseCongressCommand('m tz f')).toBe(true);
    expect('error' in parseCongressCommand('m t f')).toBe(true);
    expect('error' in parseCongressCommand('m t0 z')).toBe(true);
    expect('error' in parseCongressCommand('m t0 tz')).toBe(true);
    expect('error' in parseCongressCommand('m t0 t')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseCongressCommand('u')).toEqual({ args: ['undo'] });
    expect(parseCongressCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCongressCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseCongressCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseCongressCommand('log')).toEqual({ args: ['log'] });
    expect(parseCongressCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseCongressCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseCongressCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
