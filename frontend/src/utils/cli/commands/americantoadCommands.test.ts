import { describe, expect, it } from 'vitest';
import { parseAmericanToadCommand } from './americantoadCommands';

describe('parseAmericanToadCommand', () => {
  it('parses draw', () => {
    expect(parseAmericanToadCommand('d')).toEqual({ args: ['draw'] });
    expect(parseAmericanToadCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses reserve moves', () => {
    expect(parseAmericanToadCommand('m r f')).toEqual({
      args: ['move', { zone: 'reserve' }, { zone: 'foundation' }],
    });
    expect(parseAmericanToadCommand('m r t3')).toEqual({
      args: ['move', { zone: 'reserve' }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses waste moves', () => {
    expect(parseAmericanToadCommand('m w f')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'foundation' }],
    });
    expect(parseAmericanToadCommand('m w t2')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 2 }],
    });
  });

  it('parses tableau moves', () => {
    expect(parseAmericanToadCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation' }],
    });
    expect(parseAmericanToadCommand('m t0 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'tableau', col: 5 }],
    });
    expect(parseAmericanToadCommand('m t0.2 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseAmericanToadCommand('m')).toBe(true);
    expect('error' in parseAmericanToadCommand('m t0')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseAmericanToadCommand('m x0 f')).toBe(true);
    expect('error' in parseAmericanToadCommand('m tz f')).toBe(true);
    expect('error' in parseAmericanToadCommand('m t f')).toBe(true);
    expect('error' in parseAmericanToadCommand('m t0.z f')).toBe(true);
    expect('error' in parseAmericanToadCommand('m t0 z')).toBe(true);
    expect('error' in parseAmericanToadCommand('m t0 tz')).toBe(true);
    expect('error' in parseAmericanToadCommand('m t0 t')).toBe(true);
    expect('error' in parseAmericanToadCommand('m r z')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseAmericanToadCommand('u')).toEqual({ args: ['undo'] });
    expect(parseAmericanToadCommand('h')).toEqual({ args: ['hint'] });
    expect(parseAmericanToadCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseAmericanToadCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseAmericanToadCommand('log')).toEqual({ args: ['log'] });
    expect(parseAmericanToadCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseAmericanToadCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseAmericanToadCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
