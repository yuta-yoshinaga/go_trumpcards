import { describe, expect, it } from 'vitest';
import { parseGrandfathersClockCommand } from './grandfathersclockCommands';

describe('parseGrandfathersClockCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parseGrandfathersClockCommand('m t0 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('parses tableau-to-clock-face move', () => {
    expect(parseGrandfathersClockCommand('m t2 f7')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'foundation', col: 7 }],
    });
    expect(parseGrandfathersClockCommand('m t2 f0')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'foundation', col: 0 }],
    });
  });

  // Twelve faces can hold the same suit, so a bare "f" cannot be resolved the
  // way the one-per-suit solitaires resolve theirs.
  it('rejects a clock-face target without an index', () => {
    const r = parseGrandfathersClockCommand('m t0 f');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('index is required');
  });

  it('returns error for missing args', () => {
    expect('error' in parseGrandfathersClockCommand('m')).toBe(true);
    expect('error' in parseGrandfathersClockCommand('m t0')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseGrandfathersClockCommand('m x0 t1')).toBe(true);
    expect('error' in parseGrandfathersClockCommand('m tz t1')).toBe(true);
    expect('error' in parseGrandfathersClockCommand('m t0 z')).toBe(true);
    expect('error' in parseGrandfathersClockCommand('m t0 tz')).toBe(true);
    expect('error' in parseGrandfathersClockCommand('m t0 fz')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseGrandfathersClockCommand('u')).toEqual({ args: ['undo'] });
    expect(parseGrandfathersClockCommand('h')).toEqual({ args: ['hint'] });
    expect(parseGrandfathersClockCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseGrandfathersClockCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseGrandfathersClockCommand('log')).toEqual({ args: ['log'] });
    expect(parseGrandfathersClockCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseGrandfathersClockCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseGrandfathersClockCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
