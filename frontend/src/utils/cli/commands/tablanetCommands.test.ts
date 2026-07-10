import { describe, expect, it } from 'vitest';
import { parseTablanetCommand, TABLANET_HELP } from './tablanetCommands';

describe('parseTablanetCommand', () => {
  it('parses play with only a hand index (trail)', () => {
    expect(parseTablanetCommand('p 2')).toEqual({ args: ['play', { cardIndex: 2, tableIndices: [] }] });
    expect(parseTablanetCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0, tableIndices: [] }] });
  });

  it('parses play with capture table indices', () => {
    expect(parseTablanetCommand('p 0 1 2')).toEqual({ args: ['play', { cardIndex: 0, tableIndices: [1, 2] }] });
  });

  it('returns error for play without a hand index', () => {
    expect('error' in parseTablanetCommand('p')).toBe(true);
  });

  it('returns error for a non-numeric table index', () => {
    const result = parseTablanetCommand('p 0 x');
    expect('error' in result).toBe(true);
  });

  it('parses next and nextround', () => {
    expect(parseTablanetCommand('n')).toEqual({ args: ['next'] });
    expect(parseTablanetCommand('next')).toEqual({ args: ['next'] });
    expect(parseTablanetCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseTablanetCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint and reset', () => {
    expect(parseTablanetCommand('h')).toEqual({ args: ['hint'] });
    expect(parseTablanetCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseTablanetCommand('r')).toEqual({ args: ['reset'] });
    expect(parseTablanetCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns a suggestion error for a near-miss command', () => {
    expect('error' in parseTablanetCommand('rese')).toBe(true);
  });

  it('returns error for an unknown command', () => {
    expect('error' in parseTablanetCommand('xyz')).toBe(true);
  });

  it('exposes help text', () => {
    expect(TABLANET_HELP.length).toBeGreaterThan(0);
  });
});
