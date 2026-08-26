import { describe, expect, it } from 'vitest';
import { parseSomersetCommand } from './somersetCommands';

describe('parseSomersetCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parseSomersetCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 1 }],
    });
  });

  // The server never asked for this index: dispatchTopCardMove passes -1 in
  // place of it, so `m t0 2 t3` used to move the top card and say nothing.
  it('rejects a card index instead of silently moving the top card', () => {
    const result = parseSomersetCommand('m t0 2 t3');
    expect('args' in result).toBe(false);
    expect(result).toHaveProperty('error');
  });

  it('parses tableau-to-foundation (any)', () => {
    expect(parseSomersetCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
  });

  it('parses tableau-to-foundation specific', () => {
    expect(parseSomersetCommand('m t0 f2')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation', col: 2 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseSomersetCommand('m')).toBe(true);
    expect('error' in parseSomersetCommand('m t0')).toBe(true);
  });

  it('returns error for invalid source', () => {
    expect('error' in parseSomersetCommand('m x0 t1')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseSomersetCommand('u')).toEqual({ args: ['undo'] });
    expect(parseSomersetCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSomersetCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseSomersetCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseSomersetCommand('log')).toEqual({ args: ['log'] });
    expect(parseSomersetCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseSomersetCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseSomersetCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
