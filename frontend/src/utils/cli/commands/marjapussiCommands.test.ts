import { describe, expect, it } from 'vitest';
import { MARJAPUSSI_HELP, parseMarjapussiCommand } from './marjapussiCommands';

describe('parseMarjapussiCommand', () => {
  it('parses play (short and long)', () => {
    expect(parseMarjapussiCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseMarjapussiCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseMarjapussiCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseMarjapussiCommand('n')).toEqual({ args: ['next'] });
    expect(parseMarjapussiCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseMarjapussiCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseMarjapussiCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseMarjapussiCommand('h')).toEqual({ args: ['hint'] });
    expect(parseMarjapussiCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseMarjapussiCommand('r')).toEqual({ args: ['reset'] });
    expect(parseMarjapussiCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseMarjapussiCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseMarjapussiCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(MARJAPUSSI_HELP.length).toBeGreaterThan(0);
    expect(MARJAPUSSI_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
