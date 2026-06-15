import { describe, expect, it } from 'vitest';
import { KNOCKOUT_WHIST_HELP, parseKnockoutWhistCommand } from './knockoutWhistCommands';

describe('parseKnockoutWhistCommand', () => {
  it('parses play (short)', () => {
    expect(parseKnockoutWhistCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
  });

  it('parses play (long)', () => {
    expect(parseKnockoutWhistCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseKnockoutWhistCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseKnockoutWhistCommand('n')).toEqual({ args: ['next'] });
    expect(parseKnockoutWhistCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseKnockoutWhistCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseKnockoutWhistCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseKnockoutWhistCommand('h')).toEqual({ args: ['hint'] });
    expect(parseKnockoutWhistCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseKnockoutWhistCommand('r')).toEqual({ args: ['reset'] });
    expect(parseKnockoutWhistCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseKnockoutWhistCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseKnockoutWhistCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(KNOCKOUT_WHIST_HELP.length).toBeGreaterThan(0);
    expect(KNOCKOUT_WHIST_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
