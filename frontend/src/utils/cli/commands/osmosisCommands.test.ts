import { describe, expect, it } from 'vitest';
import { OSMOSIS_HELP, parseOsmosisCommand } from './osmosisCommands';

describe('parseOsmosisCommand', () => {
  it('parses draw', () => {
    expect(parseOsmosisCommand('d')).toEqual({ args: ['draw'] });
    expect(parseOsmosisCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses no-arg commands', () => {
    expect(parseOsmosisCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseOsmosisCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseOsmosisCommand('u')).toEqual({ args: ['undo'] });
    expect(parseOsmosisCommand('h')).toEqual({ args: ['hint'] });
    expect(parseOsmosisCommand('log')).toEqual({ args: ['log'] });
    expect(parseOsmosisCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses waste-to-foundation move (compact)', () => {
    expect(parseOsmosisCommand('m w f0')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'foundation', col: 0 }],
    });
  });

  it('parses waste-to-foundation move (spaced)', () => {
    expect(parseOsmosisCommand('m w f 2')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'foundation', col: 2 }],
    });
  });

  it('parses reserve-to-foundation move (compact)', () => {
    expect(parseOsmosisCommand('m r1 f2')).toEqual({
      args: ['move', { zone: 'reserve', col: 1 }, { zone: 'foundation', col: 2 }],
    });
  });

  it('parses reserve-to-foundation move (spaced)', () => {
    expect(parseOsmosisCommand('m r 3 f 0')).toEqual({
      args: ['move', { zone: 'reserve', col: 3 }, { zone: 'foundation', col: 0 }],
    });
  });

  it('rejects move with too few args', () => {
    expect('error' in parseOsmosisCommand('m w')).toBe(true);
  });

  it('rejects invalid source', () => {
    expect('error' in parseOsmosisCommand('m x f0')).toBe(true);
  });

  it('rejects waste move with non-numeric foundation', () => {
    expect('error' in parseOsmosisCommand('m w fx')).toBe(true);
  });

  it('rejects reserve move with non-numeric column', () => {
    expect('error' in parseOsmosisCommand('m rx f0')).toBe(true);
  });

  it('rejects reserve move with missing foundation', () => {
    expect('error' in parseOsmosisCommand('m r1 fx')).toBe(true);
  });

  it('suggests a close command', () => {
    const res = parseOsmosisCommand('drw');
    expect('error' in res).toBe(true);
  });

  it('rejects an unknown command', () => {
    const res = parseOsmosisCommand('zzz');
    expect('error' in res).toBe(true);
  });

  it('exposes help text', () => {
    expect(OSMOSIS_HELP.length).toBeGreaterThan(0);
  });
});
