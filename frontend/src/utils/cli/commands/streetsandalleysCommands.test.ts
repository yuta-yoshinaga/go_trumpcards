import { describe, expect, it } from 'vitest';
import { parseStreetsAndAlleysCommand } from './streetsandalleysCommands';

describe('parseStreetsAndAlleysCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parseStreetsAndAlleysCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses tableau move with card index', () => {
    expect(parseStreetsAndAlleysCommand('m t0 2 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses tableau-to-foundation (any)', () => {
    expect(parseStreetsAndAlleysCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation' }],
    });
  });

  it('parses tableau-to-foundation specific', () => {
    expect(parseStreetsAndAlleysCommand('m t0 f2')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation', col: 2 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseStreetsAndAlleysCommand('m')).toBe(true);
    expect('error' in parseStreetsAndAlleysCommand('m t0')).toBe(true);
  });

  it('returns error for invalid source', () => {
    expect('error' in parseStreetsAndAlleysCommand('m x0 t1')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseStreetsAndAlleysCommand('u')).toEqual({ args: ['undo'] });
    expect(parseStreetsAndAlleysCommand('h')).toEqual({ args: ['hint'] });
    expect(parseStreetsAndAlleysCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseStreetsAndAlleysCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseStreetsAndAlleysCommand('log')).toEqual({ args: ['log'] });
    expect(parseStreetsAndAlleysCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseStreetsAndAlleysCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseStreetsAndAlleysCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
