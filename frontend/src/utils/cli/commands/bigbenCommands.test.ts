import { describe, expect, it } from 'vitest';
import { parseBigBenCommand } from './bigbenCommands';

describe('parseBigBenCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parseBigBenCommand('m t0 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('parses tableau-to-clock-face move', () => {
    expect(parseBigBenCommand('m t2 f7')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'foundation', col: 7 }],
    });
    expect(parseBigBenCommand('m t2 f0')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'foundation', col: 0 }],
    });
  });

  // Twelve faces can hold the same suit, so a bare "f" cannot be resolved the
  // way the one-per-suit solitaires resolve theirs.
  it('rejects a clock-face target without an index', () => {
    const r = parseBigBenCommand('m t0 f');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('index is required');
  });

  it('returns error for missing args', () => {
    expect('error' in parseBigBenCommand('m')).toBe(true);
    expect('error' in parseBigBenCommand('m t0')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseBigBenCommand('m x0 t1')).toBe(true);
    expect('error' in parseBigBenCommand('m tz t1')).toBe(true);
    expect('error' in parseBigBenCommand('m t0 z')).toBe(true);
    expect('error' in parseBigBenCommand('m t0 tz')).toBe(true);
    expect('error' in parseBigBenCommand('m t0 fz')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseBigBenCommand('u')).toEqual({ args: ['undo'] });
    expect(parseBigBenCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBigBenCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseBigBenCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseBigBenCommand('log')).toEqual({ args: ['log'] });
    expect(parseBigBenCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseBigBenCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseBigBenCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
