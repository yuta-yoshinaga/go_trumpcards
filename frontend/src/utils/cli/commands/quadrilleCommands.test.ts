import { describe, expect, it } from 'vitest';
import { parseQuadrilleCommand, QUADRILLE_HELP } from './quadrilleCommands';

describe('parseQuadrilleCommand', () => {
  it('parses bid pass', () => {
    expect(parseQuadrilleCommand('bid pass')).toEqual({ args: ['bid', { bid: 0 }] });
    expect(parseQuadrilleCommand('b pass')).toEqual({ args: ['bid', { bid: 0 }] });
    expect(parseQuadrilleCommand('b p')).toEqual({ args: ['bid', { bid: 0 }] });
  });

  it('parses bid entrar with a trump-suit letter', () => {
    expect(parseQuadrilleCommand('bid entrar s')).toEqual({ args: ['bid', { bid: 1, trumpSuit: 1 }] });
    expect(parseQuadrilleCommand('b e c')).toEqual({ args: ['bid', { bid: 1, trumpSuit: 2 }] });
  });

  it('parses bid solo with a trump-suit letter', () => {
    expect(parseQuadrilleCommand('bid solo h')).toEqual({ args: ['bid', { bid: 2, trumpSuit: 3 }] });
    expect(parseQuadrilleCommand('b s d')).toEqual({ args: ['bid', { bid: 2, trumpSuit: 4 }] });
  });

  it('returns error for entrar/solo without a valid suit', () => {
    expect('error' in parseQuadrilleCommand('bid entrar')).toBe(true);
    expect('error' in parseQuadrilleCommand('bid solo x')).toBe(true);
  });

  it('returns error for bid without a valid argument', () => {
    expect('error' in parseQuadrilleCommand('bid')).toBe(true);
  });

  it('parses play (short and long)', () => {
    expect(parseQuadrilleCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseQuadrilleCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseQuadrilleCommand('p')).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseQuadrilleCommand('n')).toEqual({ args: ['next'] });
    expect(parseQuadrilleCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseQuadrilleCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseQuadrilleCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseQuadrilleCommand('h')).toEqual({ args: ['hint'] });
    expect(parseQuadrilleCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseQuadrilleCommand('r')).toEqual({ args: ['reset'] });
    expect(parseQuadrilleCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseQuadrilleCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseQuadrilleCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(QUADRILLE_HELP.length).toBeGreaterThan(0);
    expect(QUADRILLE_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
