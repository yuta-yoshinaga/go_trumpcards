import { describe, expect, it } from 'vitest';
import { parseColoradoCommand } from './coloradoCommands';

describe('parseColoradoCommand', () => {
  it.each([
    ['r', 'reset'],
    ['reset', 'reset'],
    ['d', 'draw'],
    ['draw', 'draw'],
    ['g', 'giveup'],
    ['giveup', 'giveup'],
    ['h', 'hint'],
    ['hint', 'hint'],
    ['u', 'undo'],
    ['undo', 'undo'],
    ['ac', 'autocomplete'],
    ['autocomplete', 'autocomplete'],
    ['l', 'log'],
    ['log', 'log'],
  ])('maps %s to %s', (input, expected) => {
    expect(parseColoradoCommand(input)).toEqual({ args: [expected] });
  });

  it('parses a waste move to a foundation, which needs no index', () => {
    expect(parseColoradoCommand('m w f')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'foundation' }],
    });
  });

  it('parses burying the waste on a pile', () => {
    expect(parseColoradoCommand('m w t 7')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', idx: 7 }],
    });
  });

  // A tableau card has one legal destination, so the zone letter is optional.
  it('parses a tableau move with and without the trailing f', () => {
    const expected = { args: ['move', { zone: 'tableau', idx: 2 }, { zone: 'foundation' }] };
    expect(parseColoradoCommand('m t 2')).toEqual(expected);
    expect(parseColoradoCommand('m t 2 f')).toEqual(expected);
  });

  it('rejects a tableau-to-tableau move, which the game does not have', () => {
    expect(parseColoradoCommand('m t 0 t 5')).toEqual({
      error: 'Usage: m t <col> (a tableau card can only go to a foundation)',
    });
  });

  it('parses filling a gap straight from the stock', () => {
    expect(parseColoradoCommand('m s t 3')).toEqual({
      args: ['move', { zone: 'stock' }, { zone: 'tableau', idx: 3 }],
    });
  });

  it('rejects the stock going anywhere but a pile', () => {
    expect(parseColoradoCommand('m s f')).toEqual({ error: 'Usage: m s t <idx>' });
  });

  it.each([
    ['m', 'Usage: m w|t|s ...'],
    ['m x', 'Usage: m w|t|s ...'],
    ['m w', 'Usage: m w f | m w t <idx>'],
    ['m w z', 'Usage: m w f | m w t <idx>'],
  ])('reports a usage error for %s', (input, error) => {
    expect(parseColoradoCommand(input)).toEqual({ error });
  });

  it.each(['m w t abc', 'm t abc', 'm s t abc'])('reports a bad index in %s', (input) => {
    const result = parseColoradoCommand(input);
    expect('error' in result).toBe(true);
  });

  it('suggests a near-miss command', () => {
    const result = parseColoradoCommand('drw');
    expect('error' in result && result.error).toContain('draw');
  });

  it('reports an unknown command', () => {
    const result = parseColoradoCommand('zzz');
    expect('error' in result && result.error).toContain('zzz');
  });
});
