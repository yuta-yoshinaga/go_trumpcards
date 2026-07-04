import { describe, expect, it } from 'vitest';
import { OMBRE_HELP, parseOmbreCommand } from './ombreCommands';

describe('parseOmbreCommand', () => {
  it('parses bid pass', () => {
    expect(parseOmbreCommand('bid pass')).toEqual({ args: ['bid', { bid: 0 }] });
    expect(parseOmbreCommand('b pass')).toEqual({ args: ['bid', { bid: 0 }] });
    expect(parseOmbreCommand('b p')).toEqual({ args: ['bid', { bid: 0 }] });
  });

  it('parses bid entrar with a trump-suit letter', () => {
    expect(parseOmbreCommand('bid entrar s')).toEqual({ args: ['bid', { bid: 1, trumpSuit: 1 }] });
    expect(parseOmbreCommand('b e c')).toEqual({ args: ['bid', { bid: 1, trumpSuit: 2 }] });
  });

  it('parses bid solo with a trump-suit letter', () => {
    expect(parseOmbreCommand('bid solo h')).toEqual({ args: ['bid', { bid: 2, trumpSuit: 3 }] });
    expect(parseOmbreCommand('b s d')).toEqual({ args: ['bid', { bid: 2, trumpSuit: 4 }] });
  });

  it('returns error for entrar/solo without a valid suit', () => {
    expect('error' in parseOmbreCommand('bid entrar')).toBe(true);
    expect('error' in parseOmbreCommand('bid solo x')).toBe(true);
  });

  it('returns error for bid without a valid argument', () => {
    expect('error' in parseOmbreCommand('bid')).toBe(true);
  });

  it('parses play (short and long)', () => {
    expect(parseOmbreCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseOmbreCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseOmbreCommand('p')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseOmbreCommand('n')).toEqual({ args: ['next'] });
    expect(parseOmbreCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseOmbreCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseOmbreCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseOmbreCommand('h')).toEqual({ args: ['hint'] });
    expect(parseOmbreCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseOmbreCommand('r')).toEqual({ args: ['reset'] });
    expect(parseOmbreCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseOmbreCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseOmbreCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(OMBRE_HELP.length).toBeGreaterThan(0);
    expect(OMBRE_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
