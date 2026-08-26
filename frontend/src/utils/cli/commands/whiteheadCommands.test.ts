import { describe, expect, it } from 'vitest';
import { parseWhiteheadCommand } from './whiteheadCommands';

describe('parseWhiteheadCommand', () => {
  it('parses draw', () => {
    expect(parseWhiteheadCommand('d')).toEqual({ args: ['draw'] });
  });

  it('parses move waste to tableau', () => {
    expect(parseWhiteheadCommand('m w t 3')).toEqual({
      args: ['move', { zone: 'waste' }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses move waste to foundation', () => {
    expect(parseWhiteheadCommand('m w f')).toEqual({ args: ['move', { zone: 'waste' }, { zone: 'foundation' }] });
  });

  it('parses move tableau to foundation', () => {
    expect(parseWhiteheadCommand('m t 2 f')).toEqual({
      args: ['move', { zone: 'tableau', col: 2 }, { zone: 'foundation' }],
    });
  });

  it('parses move tableau to tableau', () => {
    expect(parseWhiteheadCommand('m t 0 2 t 3')).toEqual({
      args: ['move', { zone: 'tableau', col: 0, cardIndex: 2 }, { zone: 'tableau', col: 3 }],
    });
  });

  it('parses giveup', () => {
    expect(parseWhiteheadCommand('g')).toEqual({ args: ['giveup'] });
  });

  it('parses autocomplete', () => {
    expect(parseWhiteheadCommand('ac')).toEqual({ args: ['autocomplete'] });
  });

  it('parses undo', () => {
    expect(parseWhiteheadCommand('u')).toEqual({ args: ['undo'] });
  });

  it('parses hint', () => {
    expect(parseWhiteheadCommand('h')).toEqual({ args: ['hint'] });
  });

  it('parses reset with draw count', () => {
    expect(parseWhiteheadCommand('r 3')).toEqual({ args: ['reset', undefined, undefined, { drawCount: 3 }] });
  });

  it('parses reset without args', () => {
    expect(parseWhiteheadCommand('r')).toEqual({ args: ['reset'] });
  });

  it('returns error for invalid move', () => {
    const result = parseWhiteheadCommand('m');
    expect('error' in result).toBe(true);
  });

  it('returns error for unknown command', () => {
    const result = parseWhiteheadCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
