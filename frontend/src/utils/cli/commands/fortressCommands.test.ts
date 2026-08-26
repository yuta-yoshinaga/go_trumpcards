import { describe, expect, it } from 'vitest';
import { parseFortressCommand } from './fortressCommands';

describe('parseFortressCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parseFortressCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 1 }],
    });
  });

  // The server never asked for this index: dispatchTopCardMove passes -1 in
  // place of it, so `m t0 2 t3` used to move the top card and say nothing.
  it('rejects a card index instead of silently moving the top card', () => {
    const result = parseFortressCommand('m t0 2 t3');
    expect('args' in result).toBe(false);
    expect(result).toHaveProperty('error');
  });

  it('parses tableau-to-foundation (any)', () => {
    expect(parseFortressCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
  });

  it('parses tableau-to-foundation specific', () => {
    expect(parseFortressCommand('m t0 f2')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation', col: 2 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseFortressCommand('m')).toBe(true);
    expect('error' in parseFortressCommand('m t0')).toBe(true);
  });

  it('returns error for invalid source', () => {
    expect('error' in parseFortressCommand('m x0 t1')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseFortressCommand('u')).toEqual({ args: ['undo'] });
    expect(parseFortressCommand('h')).toEqual({ args: ['hint'] });
    expect(parseFortressCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseFortressCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseFortressCommand('log')).toEqual({ args: ['log'] });
    expect(parseFortressCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseFortressCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseFortressCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
