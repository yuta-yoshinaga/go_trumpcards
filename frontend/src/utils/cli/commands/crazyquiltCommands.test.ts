import { describe, expect, it } from 'vitest';
import { parseCrazyQuiltCommand } from './crazyquiltCommands';

describe('parseCrazyQuiltCommand', () => {
  it('parses draw', () => {
    expect(parseCrazyQuiltCommand('d')).toEqual({ args: ['draw'] });
    expect(parseCrazyQuiltCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses a quilt card to a foundation', () => {
    expect(parseCrazyQuiltCommand('m q0 f')).toEqual({
      args: ['move', { zone: 'quilt', col: 0 }, { zone: 'foundation' }],
    });
  });

  // The quilt's main release valve: one rank away, suit irrelevant.
  it('parses a quilt card onto the waste', () => {
    expect(parseCrazyQuiltCommand('m q7 w')).toEqual({
      args: ['move', { zone: 'quilt', col: 7 }, { zone: 'waste' }],
    });
  });

  it('parses the waste to a foundation', () => {
    expect(parseCrazyQuiltCommand('m w f')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'foundation' }],
    });
  });

  // The stock is never a move source -- it is only ever turned.
  it.each(['m s f', 'm s w'])('rejects the stock as a source in %s', (input) => {
    const r = parseCrazyQuiltCommand(input);
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toBe('Invalid source: use q<cell> (quilt) or w (waste)');
  });

  // Only a quilt card goes onto the waste; the waste cannot stack on itself.
  it('rejects the waste onto the waste', () => {
    const r = parseCrazyQuiltCommand('m w w');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('only a quilt card');
  });

  // The quilt is never a destination: it only comes apart.
  it('rejects the quilt as a destination', () => {
    const r = parseCrazyQuiltCommand('m q0 q1');
    expect('error' in r).toBe(true);
    if ('error' in r) expect(r.error).toContain('Invalid target');
  });

  it('returns error for invalid sources and targets', () => {
    expect('error' in parseCrazyQuiltCommand('m x0 f')).toBe(true);
    expect('error' in parseCrazyQuiltCommand('m tz f')).toBe(true);
    expect('error' in parseCrazyQuiltCommand('m t f')).toBe(true);
    expect('error' in parseCrazyQuiltCommand('m t0 z')).toBe(true);
    expect('error' in parseCrazyQuiltCommand('m t0 tz')).toBe(true);
    expect('error' in parseCrazyQuiltCommand('m t0 t')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseCrazyQuiltCommand('u')).toEqual({ args: ['undo'] });
    expect(parseCrazyQuiltCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCrazyQuiltCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseCrazyQuiltCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseCrazyQuiltCommand('log')).toEqual({ args: ['log'] });
    expect(parseCrazyQuiltCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseCrazyQuiltCommand('movee');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseCrazyQuiltCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });
});
