import { describe, expect, it } from 'vitest';
import { parseRoyalCotillionCommand } from './royalcotillionCommands';

describe('parseRoyalCotillionCommand', () => {
  it('parses draw', () => {
    expect(parseRoyalCotillionCommand('d')).toEqual({ args: ['draw'] });
    expect(parseRoyalCotillionCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses a slot move to a foundation', () => {
    expect(parseRoyalCotillionCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
  });

  it('parses a reserve move to a foundation', () => {
    expect(parseRoyalCotillionCommand('m r1 f')).toEqual({
      args: ['move', { zone: 'reserve', col: 1 }, { zone: 'foundation' }],
    });
  });

  // A slot holds one card and refills only from the stock or the waste, so a
  // board card has nowhere to go but a foundation.
  it.each(['m t0 t5', 'm r1 t5'])('rejects %s, since a board card can only go up', (input) => {
    const r = parseRoyalCotillionCommand(input);
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('only go to a foundation');
  });

  // Negative control: the waste CAN refill a slot.
  it('parses the waste refilling a slot', () => {
    expect(parseRoyalCotillionCommand('m w t5')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 5 }],
    });
  });

  it('parses waste moves', () => {
    expect(parseRoyalCotillionCommand('m w f')).toEqual({ args: ['move', { zone: 'waste' }, { zone: 'foundation' }] });
    expect(parseRoyalCotillionCommand('m w t2')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 2 }],
    });
  });

  it('parses the stock filling a gap', () => {
    expect(parseRoyalCotillionCommand('m s t3')).toEqual({
      args: ['move', { zone: 'stock' }, { zone: 'tableau', col: 3 }],
    });
  });

  // The stock never reaches a foundation directly.
  it('rejects the stock going to a foundation', () => {
    const r = parseRoyalCotillionCommand('m s f');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('empty slot');
  });

  it('returns error for missing args', () => {
    expect('error' in parseRoyalCotillionCommand('m')).toBe(true);
    expect('error' in parseRoyalCotillionCommand('m t0')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseRoyalCotillionCommand('m x0 f')).toBe(true);
    expect('error' in parseRoyalCotillionCommand('m tz f')).toBe(true);
    expect('error' in parseRoyalCotillionCommand('m t f')).toBe(true);
    expect('error' in parseRoyalCotillionCommand('m t0 z')).toBe(true);
    expect('error' in parseRoyalCotillionCommand('m t0 tz')).toBe(true);
    expect('error' in parseRoyalCotillionCommand('m t0 t')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseRoyalCotillionCommand('u')).toEqual({ args: ['undo'] });
    expect(parseRoyalCotillionCommand('h')).toEqual({ args: ['hint'] });
    expect(parseRoyalCotillionCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseRoyalCotillionCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseRoyalCotillionCommand('log')).toEqual({ args: ['log'] });
    expect(parseRoyalCotillionCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseRoyalCotillionCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseRoyalCotillionCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
