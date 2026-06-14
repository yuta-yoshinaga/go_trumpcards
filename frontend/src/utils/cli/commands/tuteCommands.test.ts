import { describe, expect, it } from 'vitest';
import { parseTuteCommand, TUTE_HELP } from './tuteCommands';

describe('parseTuteCommand', () => {
  it('parses play (short)', () => {
    expect(parseTuteCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
  });

  it('parses play (long)', () => {
    expect(parseTuteCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseTuteCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses marriage (short)', () => {
    expect(parseTuteCommand('m 3')).toEqual({ args: ['marriage', { suit: 3 }] });
  });

  it('parses marriage (long)', () => {
    expect(parseTuteCommand('marriage 1')).toEqual({ args: ['marriage', { suit: 1 }] });
  });

  it('returns error for marriage without suit', () => {
    const result = parseTuteCommand('m');
    expect('error' in result).toBe(true);
  });

  it('parses tute', () => {
    expect(parseTuteCommand('tute')).toEqual({ args: ['tute'] });
  });

  it('parses next (short and long)', () => {
    expect(parseTuteCommand('n')).toEqual({ args: ['next'] });
    expect(parseTuteCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseTuteCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseTuteCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseTuteCommand('h')).toEqual({ args: ['hint'] });
    expect(parseTuteCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseTuteCommand('r')).toEqual({ args: ['reset'] });
    expect(parseTuteCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseTuteCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseTuteCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(TUTE_HELP.length).toBeGreaterThan(0);
    expect(TUTE_HELP.some((line) => line.includes('tute'))).toBe(true);
  });
});
