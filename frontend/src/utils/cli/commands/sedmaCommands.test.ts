import { describe, expect, it } from 'vitest';
import { parseSedmaCommand, SEDMA_HELP } from './sedmaCommands';

describe('parseSedmaCommand', () => {
  it('parses play (short)', () => {
    expect(parseSedmaCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2 }] });
  });

  it('parses play (long)', () => {
    expect(parseSedmaCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('returns error for play without index', () => {
    const result = parseSedmaCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next (short and long)', () => {
    expect(parseSedmaCommand('n')).toEqual({ args: ['next'] });
    expect(parseSedmaCommand('next')).toEqual({ args: ['next'] });
  });

  it('parses nextround (short and long)', () => {
    expect(parseSedmaCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseSedmaCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint (short and long)', () => {
    expect(parseSedmaCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSedmaCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses reset (short and long)', () => {
    expect(parseSedmaCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSedmaCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseSedmaCommand('paly 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns a plain error for an unknown command', () => {
    const result = parseSedmaCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(SEDMA_HELP.length).toBeGreaterThan(0);
    expect(SEDMA_HELP.some((line) => line.includes('Play a card'))).toBe(true);
  });
});
