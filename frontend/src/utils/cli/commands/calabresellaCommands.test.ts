import { describe, expect, it } from 'vitest';
import { CALABRESELLA_HELP, parseCalabresellaCommand } from './calabresellaCommands';

describe('parseCalabresellaCommand', () => {
  it('parses bid pass', () => {
    expect(parseCalabresellaCommand('bid pass')).toEqual({ args: ['bid', { bid: 0 }] });
    expect(parseCalabresellaCommand('b pass')).toEqual({ args: ['bid', { bid: 0 }] });
  });

  it('parses bid chiamo', () => {
    expect(parseCalabresellaCommand('bid chiamo')).toEqual({ args: ['bid', { bid: 1 }] });
    expect(parseCalabresellaCommand('b c')).toEqual({ args: ['bid', { bid: 1 }] });
  });

  it('parses bid solo', () => {
    expect(parseCalabresellaCommand('bid solo')).toEqual({ args: ['bid', { bid: 2 }] });
    expect(parseCalabresellaCommand('b s')).toEqual({ args: ['bid', { bid: 2 }] });
  });

  it('returns error for bid without a valid argument', () => {
    const result = parseCalabresellaCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('parses discard (short and long)', () => {
    expect(parseCalabresellaCommand('d 1')).toEqual({ args: ['discard', { cardIndex: 1 }] });
    expect(parseCalabresellaCommand('discard 3')).toEqual({ args: ['discard', { cardIndex: 3 }] });
  });

  it('returns error for discard without index', () => {
    const result = parseCalabresellaCommand('d');
    expect('error' in result).toBe(true);
  });

  it('parses play (short and long)', () => {
    expect(parseCalabresellaCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
    expect(parseCalabresellaCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseCalabresellaCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseCalabresellaCommand('n')).toEqual({ args: ['next'] });
    expect(parseCalabresellaCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseCalabresellaCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseCalabresellaCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseCalabresellaCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCalabresellaCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseCalabresellaCommand('r')).toEqual({ args: ['reset'] });
    expect(parseCalabresellaCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseCalabresellaCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseCalabresellaCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(CALABRESELLA_HELP.length).toBeGreaterThan(0);
    expect(CALABRESELLA_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
