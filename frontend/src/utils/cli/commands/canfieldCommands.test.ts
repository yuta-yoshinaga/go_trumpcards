import { describe, expect, it } from 'vitest';
import { parseCanfieldCommand } from './canfieldCommands';

describe('parseCanfieldCommand', () => {
  it('parses draw', () => {
    expect(parseCanfieldCommand('d')).toEqual({ args: ['draw'] });
    expect(parseCanfieldCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses waste-to-foundation move', () => {
    expect(parseCanfieldCommand('m w f')).toEqual({
      args: ['move', { zone: 'waste', col: undefined, cardIndex: undefined }, { zone: 'foundation', col: undefined }],
    });
  });

  it('parses reserve-to-tableau move', () => {
    expect(parseCanfieldCommand('m rs t2')).toEqual({
      args: ['move', { zone: 'reserve', col: undefined, cardIndex: undefined }, { zone: 'tableau', col: 2 }],
    });
  });

  it('parses tableau-to-tableau with index', () => {
    expect(parseCanfieldCommand('m t0 3 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 3 }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses tableau to specific foundation', () => {
    expect(parseCanfieldCommand('m t1 f2')).toEqual({
      args: ['move', { zone: 'tableau', col: 1, cardIndex: undefined }, { zone: 'foundation', col: 2 }],
    });
  });

  it('rejects invalid source', () => {
    expect('error' in parseCanfieldCommand('m x f')).toBe(true);
  });

  it('rejects invalid target', () => {
    expect('error' in parseCanfieldCommand('m w x')).toBe(true);
  });

  it('rejects missing target', () => {
    expect('error' in parseCanfieldCommand('m w')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseCanfieldCommand('u')).toEqual({ args: ['undo'] });
    expect(parseCanfieldCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCanfieldCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseCanfieldCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseCanfieldCommand('log')).toEqual({ args: ['log'] });
    expect(parseCanfieldCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseCanfieldCommand('draww');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseCanfieldCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
