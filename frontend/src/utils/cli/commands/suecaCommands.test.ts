import { describe, expect, it } from 'vitest';
import { parseSuecaCommand, SUECA_HELP } from './suecaCommands';

describe('parseSuecaCommand', () => {
  it('parses play (short)', () => {
    expect(parseSuecaCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
  });

  it('parses play (long)', () => {
    expect(parseSuecaCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseSuecaCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseSuecaCommand('n')).toEqual({ args: ['next'] });
    expect(parseSuecaCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseSuecaCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseSuecaCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseSuecaCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSuecaCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseSuecaCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSuecaCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseSuecaCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseSuecaCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(SUECA_HELP.length).toBeGreaterThan(0);
    expect(SUECA_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
