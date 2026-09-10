import { describe, expect, it } from 'vitest';
import { parseMatrimonyCommand } from './matrimonyCommands';

describe('parseMatrimonyCommand', () => {
  it('parses draw', () => {
    expect(parseMatrimonyCommand('d')).toEqual({ args: ['draw'] });
    expect(parseMatrimonyCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses a slot move to a foundation', () => {
    expect(parseMatrimonyCommand('m t0 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'foundation' }],
    });
  });

  // A slot holds one card and refills only from the stock or the waste, so a
  // board card has nowhere to go but a foundation.
  it.each(['m t0 t5'])('rejects %s, since a board card can only go up', (input) => {
    const r = parseMatrimonyCommand(input);
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('only go to a foundation');
  });

  // Negative control: the waste CAN refill a slot.
  it('parses the waste refilling a slot', () => {
    expect(parseMatrimonyCommand('m w t5')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 5 }],
    });
  });

  it('parses waste moves', () => {
    expect(parseMatrimonyCommand('m w f')).toEqual({ args: ['move', { zone: 'waste' }, { zone: 'foundation' }] });
    expect(parseMatrimonyCommand('m w t2')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 2 }],
    });
  });

  it('parses the stock filling a gap', () => {
    expect(parseMatrimonyCommand('m s t3')).toEqual({
      args: ['move', { zone: 'stock' }, { zone: 'tableau', col: 3 }],
    });
  });

  // The stock never reaches a foundation directly.
  it('rejects the stock going to a foundation', () => {
    const r = parseMatrimonyCommand('m s f');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('empty slot');
  });

  it('returns error for missing args', () => {
    expect('error' in parseMatrimonyCommand('m')).toBe(true);
    expect('error' in parseMatrimonyCommand('m t0')).toBe(true);
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseMatrimonyCommand('m x0 f')).toBe(true);
    expect('error' in parseMatrimonyCommand('m tz f')).toBe(true);
    expect('error' in parseMatrimonyCommand('m t f')).toBe(true);
    expect('error' in parseMatrimonyCommand('m t0 z')).toBe(true);
    expect('error' in parseMatrimonyCommand('m t0 tz')).toBe(true);
    expect('error' in parseMatrimonyCommand('m t0 t')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseMatrimonyCommand('u')).toEqual({ args: ['undo'] });
    expect(parseMatrimonyCommand('h')).toEqual({ args: ['hint'] });
    expect(parseMatrimonyCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseMatrimonyCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseMatrimonyCommand('log')).toEqual({ args: ['log'] });
    expect(parseMatrimonyCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseMatrimonyCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseMatrimonyCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
