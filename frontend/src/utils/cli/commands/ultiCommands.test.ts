import { describe, expect, it } from 'vitest';
import { parseUltiCommand, ULTI_HELP } from './ultiCommands';

describe('parseUltiCommand', () => {
  it('parses bid party with a trump-suit letter', () => {
    expect(parseUltiCommand('bid party s')).toEqual({ args: ['bid', { contract: 'party', trumpSuit: 1 }] });
    expect(parseUltiCommand('b p c')).toEqual({ args: ['bid', { contract: 'party', trumpSuit: 2 }] });
    expect(parseUltiCommand('bid party h')).toEqual({ args: ['bid', { contract: 'party', trumpSuit: 3 }] });
    expect(parseUltiCommand('bid party d')).toEqual({ args: ['bid', { contract: 'party', trumpSuit: 4 }] });
  });

  it('parses bid betli and durchmarsch without a trump', () => {
    expect(parseUltiCommand('bid betli')).toEqual({ args: ['bid', { contract: 'betli' }] });
    expect(parseUltiCommand('b b')).toEqual({ args: ['bid', { contract: 'betli' }] });
    expect(parseUltiCommand('bid durchmarsch')).toEqual({ args: ['bid', { contract: 'durchmarsch' }] });
    expect(parseUltiCommand('b d')).toEqual({ args: ['bid', { contract: 'durchmarsch' }] });
  });

  it('parses bid ulti with a trump-suit letter', () => {
    expect(parseUltiCommand('bid ulti s')).toEqual({ args: ['bid', { contract: 'ulti', trumpSuit: 1 }] });
    expect(parseUltiCommand('b u h')).toEqual({ args: ['bid', { contract: 'ulti', trumpSuit: 3 }] });
  });

  it('returns error for party or ulti without a valid suit', () => {
    expect('error' in parseUltiCommand('bid party')).toBe(true);
    expect('error' in parseUltiCommand('bid party x')).toBe(true);
    expect('error' in parseUltiCommand('bid ulti')).toBe(true);
    expect('error' in parseUltiCommand('bid ulti x')).toBe(true);
  });

  it('returns error for bid without a valid argument', () => {
    expect('error' in parseUltiCommand('bid')).toBe(true);
    expect('error' in parseUltiCommand('bid xyz')).toBe(true);
  });

  it('parses discard with two indices (short and long)', () => {
    expect(parseUltiCommand('discard 0 1')).toEqual({ args: ['discard', { cardIndices: [0, 1] }] });
    expect(parseUltiCommand('d 3 5')).toEqual({ args: ['discard', { cardIndices: [3, 5] }] });
  });

  it('returns error for discard without two indices', () => {
    expect('error' in parseUltiCommand('discard 0')).toBe(true);
    expect('error' in parseUltiCommand('d')).toBe(true);
  });

  it('parses play (short and long)', () => {
    expect(parseUltiCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseUltiCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseUltiCommand('p')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseUltiCommand('n')).toEqual({ args: ['next'] });
    expect(parseUltiCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseUltiCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseUltiCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseUltiCommand('h')).toEqual({ args: ['hint'] });
    expect(parseUltiCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseUltiCommand('r')).toEqual({ args: ['reset'] });
    expect(parseUltiCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseUltiCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseUltiCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(ULTI_HELP.length).toBeGreaterThan(0);
    expect(ULTI_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
