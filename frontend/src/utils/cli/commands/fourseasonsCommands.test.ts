import { describe, expect, it } from 'vitest';
import { parseFourSeasonsCommand } from './fourseasonsCommands';

describe('parseFourSeasonsCommand', () => {
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
    expect(parseFourSeasonsCommand(input)).toEqual({ args: [expected] });
  });

  it('parses a waste move to a corner', () => {
    expect(parseFourSeasonsCommand('m w f 2')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'foundation', idx: 2 }],
    });
  });

  it('parses a waste move to a cross pile', () => {
    expect(parseFourSeasonsCommand('m w t 3')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', idx: 3 }],
    });
  });

  it('parses a cross move to a corner, keeping the two indices distinct', () => {
    expect(parseFourSeasonsCommand('m t 4 f 1')).toEqual({
      args: ['move', { zone: 'tableau', idx: 4 }, { zone: 'foundation', idx: 1 }],
    });
  });

  it('parses a cross move to another cross pile', () => {
    expect(parseFourSeasonsCommand('m t 0 t 2')).toEqual({
      args: ['move', { zone: 'tableau', idx: 0 }, { zone: 'tableau', idx: 2 }],
    });
  });

  it('accepts the long move alias', () => {
    expect(parseFourSeasonsCommand('move w f 0')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'foundation', idx: 0 }],
    });
  });

  it.each([
    ['m', 'a bare move'],
    ['m x f 1', 'an unknown source zone'],
    ['m w', 'a waste move with no destination'],
    ['m w x 1', 'a waste move with an unknown destination'],
    ['m w f', 'a waste move with no index'],
    ['m w f abc', 'a waste move with a non-numeric index'],
    ['m t', 'a cross move with no column'],
    ['m t abc f 1', 'a cross move with a non-numeric column'],
    ['m t 0', 'a cross move with no destination'],
    ['m t 0 x 1', 'a cross move with an unknown destination'],
    ['m t 0 f', 'a cross move with no index'],
    ['m t 0 f abc', 'a cross move with a non-numeric index'],
  ])('rejects %s (%s)', (input) => {
    const result = parseFourSeasonsCommand(input);
    expect(result).toHaveProperty('error');
    expect(result).not.toHaveProperty('args');
  });

  it('rejects an unknown command with a message naming it', () => {
    const result = parseFourSeasonsCommand('frobnicate');
    expect(result).toHaveProperty('error');
    expect('error' in result && result.error).toBeTruthy();
  });
});
