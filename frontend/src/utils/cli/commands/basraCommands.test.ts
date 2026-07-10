import { describe, expect, it } from 'vitest';
import { BASRA_HELP, parseBasraCommand } from './basraCommands';

describe('parseBasraCommand', () => {
  it('parses play with only a hand index (trail)', () => {
    expect(parseBasraCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2, tableIndices: [] }] });
    expect(parseBasraCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0, tableIndices: [] }] });
  });

  it('parses play with capture table indices', () => {
    expect(parseBasraCommand('p 0 1 2')).toEqual({ args: ['play', { cardIndex: 0, tableIndices: [1, 2] }] });
  });

  it('returns error for play without a hand index', () => {
    expect('error' in parseBasraCommand('p')).toBe(true);
  });

  it('returns error for a non-numeric table index', () => {
    const result = parseBasraCommand('p 0 x');
    expect('error' in result).toBe(true);
  });

  it('parses next and nextround', () => {
    expect(parseBasraCommand('n')).toEqual({ args: ['next'] });
    expect(parseBasraCommand('next')).toEqual({ args: ['next'] });
    expect(parseBasraCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseBasraCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint and reset', () => {
    expect(parseBasraCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBasraCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseBasraCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBasraCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns a suggestion error for a near-miss command', () => {
    expect('error' in parseBasraCommand('rese')).toBe(true);
  });

  it('returns error for an unknown command', () => {
    expect('error' in parseBasraCommand('xyz')).toBe(true);
  });

  it('exposes help text', () => {
    expect(BASRA_HELP.length).toBeGreaterThan(0);
  });
});
