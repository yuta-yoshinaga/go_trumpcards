import { describe, expect, it } from 'vitest';
import { parseAgnesCommand } from './agnesCommands';

describe('parseAgnesCommand', () => {
  it('parses deal', () => {
    expect(parseAgnesCommand('d')).toEqual({ args: ['deal'] });
    expect(parseAgnesCommand('deal')).toEqual({ args: ['deal'] });
  });

  it('parses tableau-to-foundation move', () => {
    expect(parseAgnesCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation', col: undefined }],
    });
  });

  it('parses tableau-to-tableau move', () => {
    expect(parseAgnesCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses tableau-to-tableau with index', () => {
    expect(parseAgnesCommand('m t0 3 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 3 }, { zone: 'tableau', col: 1 }],
    });
  });

  it('rejects invalid source', () => {
    expect('error' in parseAgnesCommand('m w f')).toBe(true);
  });

  it('rejects invalid target', () => {
    expect('error' in parseAgnesCommand('m t0 x')).toBe(true);
  });

  it('rejects missing target', () => {
    expect('error' in parseAgnesCommand('m t0')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseAgnesCommand('u')).toEqual({ args: ['undo'] });
    expect(parseAgnesCommand('h')).toEqual({ args: ['hint'] });
    expect(parseAgnesCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseAgnesCommand('l')).toEqual({ args: ['log'] });
    expect(parseAgnesCommand('log')).toEqual({ args: ['log'] });
    expect(parseAgnesCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseAgnesCommand('deall');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseAgnesCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
