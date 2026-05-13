import { describe, expect, it } from 'vitest';
import { parseBeleagueredCastleCommand } from './beleagueredcastleCommands';

describe('parseBeleagueredCastleCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parseBeleagueredCastleCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses tableau move with card index', () => {
    expect(parseBeleagueredCastleCommand('m t0 2 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses tableau-to-foundation (any)', () => {
    expect(parseBeleagueredCastleCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation' }],
    });
  });

  it('parses tableau-to-foundation specific', () => {
    expect(parseBeleagueredCastleCommand('m t0 f2')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation', col: 2 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseBeleagueredCastleCommand('m')).toBe(true);
    expect('error' in parseBeleagueredCastleCommand('m t0')).toBe(true);
  });

  it('returns error for invalid source', () => {
    expect('error' in parseBeleagueredCastleCommand('m x0 t1')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseBeleagueredCastleCommand('u')).toEqual({ args: ['undo'] });
    expect(parseBeleagueredCastleCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBeleagueredCastleCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseBeleagueredCastleCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseBeleagueredCastleCommand('log')).toEqual({ args: ['log'] });
    expect(parseBeleagueredCastleCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseBeleagueredCastleCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseBeleagueredCastleCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
