import { describe, expect, it } from 'vitest';
import { parseSpoilFiveCommand, SPOIL_FIVE_HELP } from './spoilFiveCommands';

describe('parseSpoilFiveCommand', () => {
  it('parses play (short)', () => {
    expect(parseSpoilFiveCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
  });

  it('parses play (long)', () => {
    expect(parseSpoilFiveCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseSpoilFiveCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseSpoilFiveCommand('n')).toEqual({ args: ['next'] });
    expect(parseSpoilFiveCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseSpoilFiveCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseSpoilFiveCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseSpoilFiveCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSpoilFiveCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseSpoilFiveCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSpoilFiveCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseSpoilFiveCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseSpoilFiveCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(SPOIL_FIVE_HELP.length).toBeGreaterThan(0);
    expect(SPOIL_FIVE_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
