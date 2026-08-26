import { describe, expect, it } from 'vitest';
import { parsePerseveranceCommand } from './perseveranceCommands';

describe('parsePerseveranceCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parsePerseveranceCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses tableau move with card index', () => {
    expect(parsePerseveranceCommand('m t0 2 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses tableau-to-foundation (any)', () => {
    expect(parsePerseveranceCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation' }],
    });
  });

  it('parses tableau-to-foundation specific', () => {
    expect(parsePerseveranceCommand('m t0 f2')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation', col: 2 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parsePerseveranceCommand('m')).toBe(true);
    expect('error' in parsePerseveranceCommand('m t0')).toBe(true);
  });

  it('returns error for invalid source', () => {
    expect('error' in parsePerseveranceCommand('m x0 t1')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parsePerseveranceCommand('u')).toEqual({ args: ['undo'] });
    expect(parsePerseveranceCommand('h')).toEqual({ args: ['hint'] });
    expect(parsePerseveranceCommand('g')).toEqual({ args: ['giveup'] });
    expect(parsePerseveranceCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parsePerseveranceCommand('log')).toEqual({ args: ['log'] });
    expect(parsePerseveranceCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parsePerseveranceCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parsePerseveranceCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
