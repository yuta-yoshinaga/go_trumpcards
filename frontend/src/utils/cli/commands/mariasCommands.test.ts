import { describe, expect, it } from 'vitest';
import { MARIAS_HELP, parseMariasCommand } from './mariasCommands';

describe('parseMariasCommand', () => {
  it('parses play (short)', () => {
    expect(parseMariasCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
  });

  it('parses play (long)', () => {
    expect(parseMariasCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseMariasCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseMariasCommand('n')).toEqual({ args: ['next'] });
    expect(parseMariasCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseMariasCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseMariasCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseMariasCommand('h')).toEqual({ args: ['hint'] });
    expect(parseMariasCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseMariasCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMariasCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseMariasCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseMariasCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(MARIAS_HELP.length).toBeGreaterThan(0);
    expect(MARIAS_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
