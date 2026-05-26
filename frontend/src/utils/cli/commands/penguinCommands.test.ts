import { describe, expect, it } from 'vitest';
import { parsePenguinCommand } from './penguinCommands';

describe('parsePenguinCommand', () => {
  it('parses tableau-to-tableau move', () => {
    expect(parsePenguinCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 1 }],
    });
  });

  it('returns error for card index format (3-arg form)', () => {
    // The implementation requires args.length >= 4 for the card index form,
    // so 'm t0 2 t3' (3 args after cmd) falls through to the to-zone check
    const result = parsePenguinCommand('m t0 2 t3');
    expect('error' in result).toBe(true);
  });

  it('parses tableau-to-foundation', () => {
    expect(parsePenguinCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
  });

  it('parses tableau-to-freecell', () => {
    expect(parsePenguinCommand('m t0 c4')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'freecell', cell: 4 }],
    });
  });

  it('parses freecell-to-tableau', () => {
    expect(parsePenguinCommand('m c0 t1')).toEqual({
      args: ['move', { zone: 'freecell', cell: 0 }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses freecell-to-foundation', () => {
    expect(parsePenguinCommand('m c0 f')).toEqual({
      args: ['move', { zone: 'freecell', cell: 0 }, { zone: 'foundation' }],
    });
  });

  it('returns error for missing move args', () => {
    expect('error' in parsePenguinCommand('m')).toBe(true);
    expect('error' in parsePenguinCommand('m t0')).toBe(true);
  });

  it('returns error for invalid source', () => {
    expect('error' in parsePenguinCommand('m x0 t1')).toBe(true);
  });

  it('returns error for invalid freecell target', () => {
    const result = parsePenguinCommand('m c0 x1');
    expect('error' in result).toBe(true);
  });

  it('returns error for invalid tableau target', () => {
    const result = parsePenguinCommand('m t0 x1');
    expect('error' in result).toBe(true);
  });

  it('parses control commands', () => {
    expect(parsePenguinCommand('u')).toEqual({ args: ['undo'] });
    expect(parsePenguinCommand('undo')).toEqual({ args: ['undo'] });
    expect(parsePenguinCommand('h')).toEqual({ args: ['hint'] });
    expect(parsePenguinCommand('hint')).toEqual({ args: ['hint'] });
    expect(parsePenguinCommand('g')).toEqual({ args: ['giveup'] });
    expect(parsePenguinCommand('giveup')).toEqual({ args: ['giveup'] });
    expect(parsePenguinCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parsePenguinCommand('autocomplete')).toEqual({ args: ['autocomplete'] });
    expect(parsePenguinCommand('log')).toEqual({ args: ['log'] });
    expect(parsePenguinCommand('r')).toEqual({ args: ['reset'] });
    expect(parsePenguinCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parsePenguinCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parsePenguinCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
