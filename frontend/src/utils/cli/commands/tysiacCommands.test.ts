import { describe, expect, it } from 'vitest';
import { parseTysiacCommand, TYSIAC_HELP } from './tysiacCommands';

describe('parseTysiacCommand', () => {
  it('parses bid raise', () => {
    expect(parseTysiacCommand('bid raise')).toEqual({ args: ['bid', { raise: true }] });
    expect(parseTysiacCommand('b raise')).toEqual({ args: ['bid', { raise: true }] });
  });

  it('parses bid pass', () => {
    expect(parseTysiacCommand('bid pass')).toEqual({ args: ['bid', { raise: false }] });
    expect(parseTysiacCommand('b pass')).toEqual({ args: ['bid', { raise: false }] });
  });

  it('returns error for bid without a valid argument', () => {
    const result = parseTysiacCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('parses discard (short and long)', () => {
    expect(parseTysiacCommand('d 1')).toEqual({ args: ['discard', { cardIndex: 1 }] });
    expect(parseTysiacCommand('discard 3')).toEqual({ args: ['discard', { cardIndex: 3 }] });
  });

  it('returns error for discard without index', () => {
    const result = parseTysiacCommand('d');
    expect('error' in result).toBe(true);
  });

  it('parses play (short and long)', () => {
    expect(parseTysiacCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseTysiacCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseTysiacCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseTysiacCommand('n')).toEqual({ args: ['next'] });
    expect(parseTysiacCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseTysiacCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseTysiacCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseTysiacCommand('h')).toEqual({ args: ['hint'] });
    expect(parseTysiacCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseTysiacCommand('r')).toEqual({ args: ['reset'] });
    expect(parseTysiacCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseTysiacCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseTysiacCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(TYSIAC_HELP.length).toBeGreaterThan(0);
    expect(TYSIAC_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
