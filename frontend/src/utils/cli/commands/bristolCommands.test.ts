import { describe, expect, it } from 'vitest';
import { BRISTOL_HELP, parseBristolCommand } from './bristolCommands';

describe('parseBristolCommand', () => {
  it('parses draw', () => {
    expect(parseBristolCommand('d')).toEqual({ args: ['draw'] });
    expect(parseBristolCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses no-arg commands', () => {
    expect(parseBristolCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseBristolCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseBristolCommand('u')).toEqual({ args: ['undo'] });
    expect(parseBristolCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBristolCommand('log')).toEqual({ args: ['log'] });
    expect(parseBristolCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses tableau-to-tableau move (compact)', () => {
    expect(parseBristolCommand('m t0 t1')).toEqual({
      args: ['move', { zone: 'tableau', col: 0 }, { zone: 'tableau', col: 1 }],
    });
  });

  it('parses tableau-to-tableau move (spaced)', () => {
    expect(parseBristolCommand('m t 2 t 5')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('parses tableau-to-foundation move', () => {
    expect(parseBristolCommand('m t3 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 3 }, { zone: 'foundation' }],
    });
  });

  it('parses fan-to-tableau move (compact)', () => {
    expect(parseBristolCommand('m n1 t4')).toEqual({
      args: ['move', { zone: 'fan', col: 1 }, { zone: 'tableau', col: 4 }],
    });
  });

  it('parses fan-to-foundation move (spaced)', () => {
    expect(parseBristolCommand('m n 2 f')).toEqual({
      args: ['move', { zone: 'fan', col: 2 }, { zone: 'foundation' }],
    });
  });

  it('rejects move with too few args', () => {
    expect('error' in parseBristolCommand('m t')).toBe(true);
  });

  it('rejects invalid source', () => {
    expect('error' in parseBristolCommand('m x f')).toBe(true);
  });

  it('rejects tableau move with non-numeric column', () => {
    expect('error' in parseBristolCommand('m tx f')).toBe(true);
  });

  it('rejects tableau-to-tableau with non-numeric destination', () => {
    expect('error' in parseBristolCommand('m t0 tx')).toBe(true);
  });

  it('rejects move with invalid destination', () => {
    expect('error' in parseBristolCommand('m t0 x')).toBe(true);
  });

  it('rejects an unknown command', () => {
    expect('error' in parseBristolCommand('zzz')).toBe(true);
  });

  it('exposes help text', () => {
    expect(BRISTOL_HELP.length).toBeGreaterThan(0);
  });
});
