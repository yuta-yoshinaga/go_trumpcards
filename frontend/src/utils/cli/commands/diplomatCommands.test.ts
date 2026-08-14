import { describe, expect, it } from 'vitest';
import { parseDiplomatCommand } from './diplomatCommands';

describe('parseDiplomatCommand', () => {
  it('parses draw', () => {
    expect(parseDiplomatCommand('d')).toEqual({ args: ['draw'] });
    expect(parseDiplomatCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses tableau moves', () => {
    expect(parseDiplomatCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
    expect(parseDiplomatCommand('m t0 t5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('parses waste moves', () => {
    expect(parseDiplomatCommand('m w f')).toEqual({ args: ['move', { zone: 'waste' }, { zone: 'foundation' }] });
    expect(parseDiplomatCommand('m w t2')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 2 }],
    });
  });

  // The stock is not a source at all here: an empty column is filled from
  // another column or the waste, never straight from the deck.
  it.each(['m s t3', 'm s f'])('rejects the stock as a source in %s', (input) => {
    const r = parseDiplomatCommand(input);
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toBe('Invalid source: use t<pile> (tableau) or w (waste)');
  });

  // Negative control: a column IS a valid source for the same destination.
  it('accepts a column moving into an empty column', () => {
    expect(parseDiplomatCommand('m t0 t3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('returns error for missing args', () => {
    expect('error' in parseDiplomatCommand('m')).toBe(true);
    expect('error' in parseDiplomatCommand('m t0')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseDiplomatCommand('m x0 f')).toBe(true);
    expect('error' in parseDiplomatCommand('m tz f')).toBe(true);
    expect('error' in parseDiplomatCommand('m t f')).toBe(true);
    expect('error' in parseDiplomatCommand('m t0 z')).toBe(true);
    expect('error' in parseDiplomatCommand('m t0 tz')).toBe(true);
    expect('error' in parseDiplomatCommand('m t0 t')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseDiplomatCommand('u')).toEqual({ args: ['undo'] });
    expect(parseDiplomatCommand('h')).toEqual({ args: ['hint'] });
    expect(parseDiplomatCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseDiplomatCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseDiplomatCommand('log')).toEqual({ args: ['log'] });
    expect(parseDiplomatCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseDiplomatCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseDiplomatCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
