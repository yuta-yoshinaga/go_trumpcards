import { describe, expect, it } from 'vitest';
import { parseFlowerGardenCommand } from './flowergardenCommands';

describe('parseFlowerGardenCommand', () => {
  it('parses bed-to-bed move', () => {
    expect(parseFlowerGardenCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses bed move with card index', () => {
    expect(parseFlowerGardenCommand('m t0 2 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses bed-to-foundation (any)', () => {
    expect(parseFlowerGardenCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation' }],
    });
  });

  it('parses bed-to-foundation specific', () => {
    expect(parseFlowerGardenCommand('m t0 f2')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: undefined }, { zone: 'foundation', col: 2 }],
    });
  });

  it('parses reserve-to-bed move', () => {
    expect(parseFlowerGardenCommand('m r0 t1')).toEqual({
      args: ['move', { zone: 'reserve', col: 0 }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses reserve-to-foundation (any)', () => {
    expect(parseFlowerGardenCommand('m r3 f')).toEqual({
      args: ['move', { zone: 'reserve', col: 3 }, { zone: 'foundation' }],
    });
  });

  it('parses reserve-to-foundation specific', () => {
    expect(parseFlowerGardenCommand('m r3 f2')).toEqual({
      args: ['move', { zone: 'reserve', col: 3 }, { zone: 'foundation', col: 2 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseFlowerGardenCommand('m')).toBe(true);
    expect('error' in parseFlowerGardenCommand('m t0')).toBe(true);
  });

  it('returns error for invalid source', () => {
    expect('error' in parseFlowerGardenCommand('m x0 t1')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseFlowerGardenCommand('u')).toEqual({ args: ['undo'] });
    expect(parseFlowerGardenCommand('h')).toEqual({ args: ['hint'] });
    expect(parseFlowerGardenCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseFlowerGardenCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseFlowerGardenCommand('log')).toEqual({ args: ['log'] });
    expect(parseFlowerGardenCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseFlowerGardenCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseFlowerGardenCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
