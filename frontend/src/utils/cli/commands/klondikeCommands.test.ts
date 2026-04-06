import { describe, expect, it } from 'vitest';
import { parseKlondikeCommand } from './klondikeCommands';

describe('parseKlondikeCommand', () => {
  it('parses draw', () => {
    expect(parseKlondikeCommand('d')).toEqual({ args: ['draw'] });
  });

  it('parses move waste to tableau', () => {
    expect(parseKlondikeCommand('m w t 3')).toEqual({ args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 3 }] });
  });

  it('parses move waste to foundation', () => {
    expect(parseKlondikeCommand('m w f')).toEqual({ args: ['move', { zone: 'waste' }, { zone: 'foundation' }] });
  });

  it('parses move tableau to foundation', () => {
    expect(parseKlondikeCommand('m t 2 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'foundation' }],
    });
  });

  it('parses move tableau to tableau', () => {
    expect(parseKlondikeCommand('m t 0 2 t 3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses giveup', () => {
    expect(parseKlondikeCommand('g')).toEqual({ args: ['giveup'] });
  });

  it('parses autocomplete', () => {
    expect(parseKlondikeCommand('ac')).toEqual({ args: ['autocomplete'] });
  });

  it('parses undo', () => {
    expect(parseKlondikeCommand('u')).toEqual({ args: ['undo'] });
  });

  it('parses hint', () => {
    expect(parseKlondikeCommand('h')).toEqual({ args: ['hint'] });
  });

  it('parses reset with draw count', () => {
    expect(parseKlondikeCommand('r 3')).toEqual({ args: ['reset', undefined, undefined, { drawCount: 3 }] });
  });

  it('parses reset without args', () => {
    expect(parseKlondikeCommand('r')).toEqual({ args: ['reset'] });
  });

  it('returns error for invalid move', () => {
    const result = parseKlondikeCommand('m');
    expect('error' in result).toBe(true);
  });

  it('returns error for unknown command', () => {
    const result = parseKlondikeCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
