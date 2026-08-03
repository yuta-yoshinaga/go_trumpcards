import { describe, expect, it } from 'vitest';
import { parseWindmillCommand } from './windmillCommands';

describe('parseWindmillCommand', () => {
  it('parses draw', () => {
    expect(parseWindmillCommand('d')).toEqual({ args: ['draw'] });
    expect(parseWindmillCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses sail moves', () => {
    expect(parseWindmillCommand('m s0 c')).toEqual({
      args: ['move', { zone: 'sail', col: 0 }, { zone: 'center' }],
    });
    expect(parseWindmillCommand('m s7 k1')).toEqual({
      args: ['move', { zone: 'sail', col: 7 }, { zone: 'corner', col: 1 }],
    });
  });

  it('parses waste moves', () => {
    expect(parseWindmillCommand('m w c')).toEqual({ args: ['move', { zone: 'waste' }, { zone: 'center' }] });
    expect(parseWindmillCommand('m w k2')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'corner', col: 2 }],
    });
  });

  it('parses the corner pull-back', () => {
    expect(parseWindmillCommand('m k0 c')).toEqual({
      args: ['move', { zone: 'corner', col: 0 }, { zone: 'center' }],
    });
  });

  // The rescue runs one way only, so a corner is never both source and target.
  it('rejects a corner-to-corner move', () => {
    const r = parseWindmillCommand('m k0 k1');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('centre');
  });

  it('returns error for missing args', () => {
    expect('error' in parseWindmillCommand('m')).toBe(true);
    expect('error' in parseWindmillCommand('m s0')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseWindmillCommand('m x0 c')).toBe(true);
    expect('error' in parseWindmillCommand('m s c')).toBe(true);
    expect('error' in parseWindmillCommand('m sz c')).toBe(true);
    expect('error' in parseWindmillCommand('m kz c')).toBe(true);
    expect('error' in parseWindmillCommand('m k c')).toBe(true);
    expect('error' in parseWindmillCommand('m s0 z')).toBe(true);
    expect('error' in parseWindmillCommand('m s0 kz')).toBe(true);
    expect('error' in parseWindmillCommand('m s0 k')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseWindmillCommand('u')).toEqual({ args: ['undo'] });
    expect(parseWindmillCommand('h')).toEqual({ args: ['hint'] });
    expect(parseWindmillCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseWindmillCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseWindmillCommand('log')).toEqual({ args: ['log'] });
    expect(parseWindmillCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseWindmillCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseWindmillCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
