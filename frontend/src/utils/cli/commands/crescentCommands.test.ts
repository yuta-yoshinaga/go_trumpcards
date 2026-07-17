import { describe, expect, it } from 'vitest';
import { parseCrescentCommand } from './crescentCommands';

describe('parseCrescentCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parseCrescentCommand('m t0 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses tableau-to-any-foundation move', () => {
    expect(parseCrescentCommand('m t2 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'foundation' }],
    });
  });

  it('parses tableau-to-specific-foundation move', () => {
    expect(parseCrescentCommand('m t2 f5')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'foundation', col: 5 }],
    });
  });

  it('errors on too few move args or bad zones', () => {
    expect('error' in parseCrescentCommand('m t0')).toBe(true);
    expect('error' in parseCrescentCommand('m x0 t1')).toBe(true);
    expect('error' in parseCrescentCommand('m t0 x1')).toBe(true);
    expect('error' in parseCrescentCommand('m tA t1')).toBe(true);
  });

  it('parses redeal', () => {
    expect(parseCrescentCommand('d')).toEqual({ args: ['redeal'] });
    expect(parseCrescentCommand('redeal')).toEqual({ args: ['redeal'] });
  });

  it('parses undo, hint, autocomplete, giveup, log, reset', () => {
    expect(parseCrescentCommand('u')).toEqual({ args: ['undo'] });
    expect(parseCrescentCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCrescentCommand('a')).toEqual({ args: ['autocomplete'] });
    expect(parseCrescentCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseCrescentCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseCrescentCommand('log')).toEqual({ args: ['log'] });
    expect(parseCrescentCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for a close typo', () => {
    const result = parseCrescentCommand('reset ');
    expect(result).toEqual({ args: ['reset'] });
    const typo = parseCrescentCommand('hnt');
    expect('error' in typo).toBe(true);
    if ('error' in typo) expect(typo.error).toContain('hint');
  });

  it('errors on an unknown command', () => {
    expect('error' in parseCrescentCommand('xyz')).toBe(true);
  });
});
