import { describe, expect, it } from 'vitest';
import { easthavenHelp, parseEasthavenCommand } from './easthavenCommands';

describe('parseEasthavenCommand', () => {
  it('parses simple commands and their aliases', () => {
    expect(parseEasthavenCommand('r')).toEqual({ args: ['reset'] });
    expect(parseEasthavenCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseEasthavenCommand('d')).toEqual({ args: ['deal'] });
    expect(parseEasthavenCommand('deal')).toEqual({ args: ['deal'] });
    expect(parseEasthavenCommand('g')).toEqual({ args: ['giveup'] });
    expect(parseEasthavenCommand('giveup')).toEqual({ args: ['giveup'] });
    expect(parseEasthavenCommand('h')).toEqual({ args: ['hint'] });
    expect(parseEasthavenCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseEasthavenCommand('ac')).toEqual({ args: ['autocomplete'] });
    expect(parseEasthavenCommand('autocomplete')).toEqual({ args: ['autocomplete'] });
    expect(parseEasthavenCommand('u')).toEqual({ args: ['undo'] });
    expect(parseEasthavenCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses m <from> <to> as a tableau top-card move', () => {
    expect(parseEasthavenCommand('m 1 4')).toEqual({
      args: ['move', { zone: 'tableau', col: 1, cardIndex: -1 }, { zone: 'tableau', col: 4 }],
    });
  });

  it('parses m t <col> f as a tableau-to-foundation move', () => {
    expect(parseEasthavenCommand('m t 2 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 2, cardIndex: -1 }, { zone: 'foundation' }],
    });
  });

  it('parses m t <col> <idx> t <col> as a run move by index', () => {
    expect(parseEasthavenCommand('m t 0 3 t 5')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 3 }, { zone: 'tableau', col: 5 }],
    });
  });

  it('is case-insensitive on the command word', () => {
    expect(parseEasthavenCommand('MOVE 1 4')).toEqual({
      args: ['move', { zone: 'tableau', col: 1, cardIndex: -1 }, { zone: 'tableau', col: 4 }],
    });
  });

  it('returns a localized error for a non-numeric column in a top-card move', () => {
    const result = parseEasthavenCommand('m x 4');
    expect('error' in result && result.error).toBe('無効な列番号です');
  });

  it('returns a localized error for a non-numeric column in a foundation move', () => {
    const result = parseEasthavenCommand('m t x f');
    expect('error' in result && result.error).toBe('無効な列番号です');
  });

  it('returns a localized error for a non-numeric arg in a run move', () => {
    const result = parseEasthavenCommand('m t 0 x t 5');
    expect('error' in result && result.error).toBe('無効な引数です');
  });

  it('returns the usage error for a malformed move', () => {
    const result = parseEasthavenCommand('m');
    expect('error' in result && typeof result.error === 'string' && result.error.startsWith('使い方')).toBe(true);
  });

  it('returns an unknown-command error including the command', () => {
    const result = parseEasthavenCommand('zzz');
    expect('error' in result && result.error).toBe('不明なコマンド: zzz');
  });
});

describe('easthavenHelp', () => {
  it('returns a non-empty list of localized help lines', () => {
    const help = easthavenHelp();
    expect(Array.isArray(help)).toBe(true);
    expect(help.length).toBeGreaterThan(0);
    expect(help.some((line) => line.includes('m t'))).toBe(true);
  });
});
