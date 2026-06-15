import { describe, expect, it } from 'vitest';
import { KLAVERJAS_HELP, parseKlaverjasCommand } from './klaverjasCommands';

describe('parseKlaverjasCommand', () => {
  it('parses play (short)', () => {
    expect(parseKlaverjasCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
  });

  it('parses play (long)', () => {
    expect(parseKlaverjasCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseKlaverjasCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseKlaverjasCommand('n')).toEqual({ args: ['next'] });
    expect(parseKlaverjasCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseKlaverjasCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseKlaverjasCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseKlaverjasCommand('h')).toEqual({ args: ['hint'] });
    expect(parseKlaverjasCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseKlaverjasCommand('r')).toEqual({ args: ['reset'] });
    expect(parseKlaverjasCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseKlaverjasCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseKlaverjasCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(KLAVERJAS_HELP.length).toBeGreaterThan(0);
    expect(KLAVERJAS_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
