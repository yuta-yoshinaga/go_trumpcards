import { describe, expect, it } from 'vitest';
import { parseBisleyCommand } from './bisleyCommands';

describe('parseBisleyCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parseBisleyCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses tableau-to-ascending-foundation move', () => {
    expect(parseBisleyCommand('m t7 a')).toEqual({
      args: ['move', { zone: 'tableau', col: 7 }, { zone: 'ace' }],
    });
  });

  it('parses tableau-to-descending-foundation move', () => {
    expect(parseBisleyCommand('m t12 k')).toEqual({
      args: ['move', { zone: 'tableau', col: 12 }, { zone: 'king' }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseBisleyCommand('m')).toBe(true);
    expect('error' in parseBisleyCommand('m t0')).toBe(true);
  });

  it('returns error for invalid source', () => {
    expect('error' in parseBisleyCommand('m x0 t1')).toBe(true);
    expect('error' in parseBisleyCommand('m tz t1')).toBe(true);
  });

  it('returns error for invalid target', () => {
    expect('error' in parseBisleyCommand('m t0 f')).toBe(true);
    expect('error' in parseBisleyCommand('m t0 tz')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseBisleyCommand('u')).toEqual({ args: ['undo'] });
    expect(parseBisleyCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBisleyCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseBisleyCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseBisleyCommand('log')).toEqual({ args: ['log'] });
    expect(parseBisleyCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseBisleyCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseBisleyCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
