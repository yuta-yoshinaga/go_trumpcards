import { describe, expect, it } from 'vitest';
import { MANILLE_HELP, parseManilleCommand } from './manilleCommands';

describe('parseManilleCommand', () => {
  it('parses play (short)', () => {
    expect(parseManilleCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
  });

  it('parses play (long)', () => {
    expect(parseManilleCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseManilleCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseManilleCommand('n')).toEqual({ args: ['next'] });
    expect(parseManilleCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseManilleCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseManilleCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseManilleCommand('h')).toEqual({ args: ['hint'] });
    expect(parseManilleCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseManilleCommand('r')).toEqual({ args: ['reset'] });
    expect(parseManilleCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseManilleCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseManilleCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(MANILLE_HELP.length).toBeGreaterThan(0);
    expect(MANILLE_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
