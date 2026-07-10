import { describe, expect, it } from 'vitest';
import { parseKingAlbertCommand } from './kingalbertCommands';

describe('parseKingAlbertCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parseKingAlbertCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses tableau move with card index', () => {
    expect(parseKingAlbertCommand('m t0 2 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses tableau-to-foundation (any)', () => {
    expect(parseKingAlbertCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation' }],
    });
  });

  it('parses tableau-to-foundation specific', () => {
    expect(parseKingAlbertCommand('m t0 f2')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation', col: 2 }],
    });
  });

  it('parses reserve-to-tableau move', () => {
    expect(parseKingAlbertCommand('m r0 t1')).toEqual({
      args: ['move', { zone: 'reserve', col: 0 }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses reserve-to-foundation (any)', () => {
    expect(parseKingAlbertCommand('m r3 f')).toEqual({
      args: ['move', { zone: 'reserve', col: 3 }, { zone: 'foundation' }],
    });
  });

  it('parses reserve-to-foundation specific', () => {
    expect(parseKingAlbertCommand('m r3 f2')).toEqual({
      args: ['move', { zone: 'reserve', col: 3 }, { zone: 'foundation', col: 2 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseKingAlbertCommand('m')).toBe(true);
    expect('error' in parseKingAlbertCommand('m t0')).toBe(true);
  });

  it('returns error for invalid source', () => {
    expect('error' in parseKingAlbertCommand('m x0 t1')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseKingAlbertCommand('u')).toEqual({ args: ['undo'] });
    expect(parseKingAlbertCommand('h')).toEqual({ args: ['hint'] });
    expect(parseKingAlbertCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseKingAlbertCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseKingAlbertCommand('log')).toEqual({ args: ['log'] });
    expect(parseKingAlbertCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseKingAlbertCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseKingAlbertCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
