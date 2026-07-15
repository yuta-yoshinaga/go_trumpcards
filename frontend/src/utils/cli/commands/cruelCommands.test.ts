import { describe, expect, it } from 'vitest';
import { cruelHelp, parseCruelCommand } from './cruelCommands';

describe('parseCruelCommand', () => {
  it('parses simple commands and their aliases', () => {
    expect(parseCruelCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCruelCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseCruelCommand('s')).toEqual({ args: ['shift'] });
    expect(parseCruelCommand('shift')).toEqual({ args: ['shift'] });
    expect(parseCruelCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseCruelCommand('giveup')).toEqual({ args: ['giveup'] });
    expect(parseCruelCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCruelCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseCruelCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseCruelCommand('autocomplete')).toEqual({ args: ['autocomplete'] });
    expect(parseCruelCommand('u')).toEqual({ args: ['undo'] });
    expect(parseCruelCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses m <from> <to> as a tableau move', () => {
    expect(parseCruelCommand('m 1 4')).toEqual({
      args: ['move', { zone: 'tableau', col: 1 }, { zone: 'tableau', col: 4 }],
    });
  });

  it('parses m <from> f as a tableau-to-foundation move', () => {
    expect(parseCruelCommand('m 2 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'foundation' }],
    });
  });

  it('is case-insensitive on the command word', () => {
    expect(parseCruelCommand('MOVE 1 4')).toEqual({
      args: ['move', { zone: 'tableau', col: 1 }, { zone: 'tableau', col: 4 }],
    });
  });

  it('returns a localized error for a non-numeric source column', () => {
    const result = parseCruelCommand('m x 4');
    expect('error' in result && result.error).toBe('無効な移動元の列番号です');
  });

  it('returns a localized error for a non-numeric destination', () => {
    const result = parseCruelCommand('m 1 x');
    expect('error' in result && result.error).toBe('無効な移動先です');
  });

  it('returns the usage error for a malformed move', () => {
    const result = parseCruelCommand('m 1');
    expect('error' in result && typeof result.error === 'string' && result.error.startsWith('使い方')).toBe(true);
  });

  it('returns an unknown-command error including the command', () => {
    const result = parseCruelCommand('zzz');
    expect('error' in result && result.error).toBe('不明なコマンド: zzz');
  });
});

describe('cruelHelp', () => {
  it('returns a non-empty list of localized help lines', () => {
    const help = cruelHelp();
    expect(Array.isArray(help)).toBe(true);
    expect(help.length).toBeGreaterThan(0);
    expect(help.some((line) => line.includes('s '))).toBe(true);
  });
});
