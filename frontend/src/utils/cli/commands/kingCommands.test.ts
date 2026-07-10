import { describe, expect, it } from 'vitest';
import { KING_HELP, parseKingCommand } from './kingCommands';

describe('parseKingCommand', () => {
  it('parses contract without a trump (defaults trumpSuit to -1)', () => {
    expect(parseKingCommand('c 0')).toEqual({ args: ['contract', { contract: 0, trumpSuit: -1 }] });
    expect(parseKingCommand('contract 4')).toEqual({ args: ['contract', { contract: 4, trumpSuit: -1 }] });
  });

  it('parses the King (Trump) contract with an explicit trump suit', () => {
    expect(parseKingCommand('c 6 3')).toEqual({ args: ['contract', { contract: 6, trumpSuit: 3 }] });
  });

  it('returns error for contract without an index', () => {
    const result = parseKingCommand('c');
    expect('error' in result).toBe(true);
  });

  it('parses play (short and long)', () => {
    expect(parseKingCommand('p 2')).toEqual({ args: ['play', { handIndex: 2 }] });
    expect(parseKingCommand('play 0')).toEqual({ args: ['play', { handIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseKingCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseKingCommand('n')).toEqual({ args: ['next'] });
    expect(parseKingCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseKingCommand('h')).toEqual({ args: ['hint'] });
    expect(parseKingCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseKingCommand('r')).toEqual({ args: ['reset'] });
    expect(parseKingCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseKingCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseKingCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(KING_HELP.length).toBeGreaterThan(0);
    expect(KING_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
