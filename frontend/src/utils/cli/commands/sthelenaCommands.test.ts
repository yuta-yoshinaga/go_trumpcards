import { describe, expect, it } from 'vitest';
import { parseStHelenaCommand } from './sthelenaCommands';

describe('parseStHelenaCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parseStHelenaCommand('m t0 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses tableau-to-any-foundation move', () => {
    expect(parseStHelenaCommand('m t2 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'foundation' }],
    });
  });

  it('parses tableau-to-specific-foundation move', () => {
    expect(parseStHelenaCommand('m t2 f5')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'foundation', col: 5 }],
    });
  });

  it('errors on too few move args or bad zones', () => {
    expect('error' in parseStHelenaCommand('m t0')).toBe(true);
    expect('error' in parseStHelenaCommand('m x0 t1')).toBe(true);
    expect('error' in parseStHelenaCommand('m t0 x1')).toBe(true);
    expect('error' in parseStHelenaCommand('m tA t1')).toBe(true);
  });

  it('parses redeal', () => {
    expect(parseStHelenaCommand('d')).toEqual({ args: ['redeal'] });
    expect(parseStHelenaCommand('redeal')).toEqual({ args: ['redeal'] });
  });

  it('parses undo, hint, autocomplete, giveup, log, reset', () => {
    expect(parseStHelenaCommand('u')).toEqual({ args: ['undo'] });
    expect(parseStHelenaCommand('h')).toEqual({ args: ['hint'] });
    expect(parseStHelenaCommand('a')).toEqual({ args: ['autocomplete'] });
    expect(parseStHelenaCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseStHelenaCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseStHelenaCommand('log')).toEqual({ args: ['log'] });
    expect(parseStHelenaCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for a close typo', () => {
    const result = parseStHelenaCommand('reset ');
    expect(result).toEqual({ args: ['reset'] });
    const typo = parseStHelenaCommand('hnt');
    expect('error' in typo).toBe(true);
    if ('error' in typo) expect(typo.error).toContain('hint');
  });

  it('errors on an unknown command', () => {
    expect('error' in parseStHelenaCommand('xyz')).toBe(true);
  });
});
