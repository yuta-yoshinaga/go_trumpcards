import { describe, expect, it } from 'vitest';
import { ALLFOURS_HELP, parseAllFoursCommand } from './allfoursCommands';

describe('parseAllFoursCommand', () => {
  it('parses stand and beg', () => {
    expect(parseAllFoursCommand('stand')).toEqual({ args: ['beg', false] });
    expect(parseAllFoursCommand('st')).toEqual({ args: ['beg', false] });
    expect(parseAllFoursCommand('beg')).toEqual({ args: ['beg', true] });
    expect(parseAllFoursCommand('bg')).toEqual({ args: ['beg', true] });
  });

  it('parses gift and run', () => {
    expect(parseAllFoursCommand('gift')).toEqual({ args: ['respond', undefined, false] });
    expect(parseAllFoursCommand('g')).toEqual({ args: ['respond', undefined, false] });
    expect(parseAllFoursCommand('run')).toEqual({ args: ['respond', undefined, true] });
  });

  it('parses play with index', () => {
    expect(parseAllFoursCommand('p 2')).toEqual({ args: ['play', undefined, undefined, 2] });
    expect(parseAllFoursCommand('play 0')).toEqual({ args: ['play', undefined, undefined, 0] });
  });

  it('returns error for play without index', () => {
    const result = parseAllFoursCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next and nextround', () => {
    expect(parseAllFoursCommand('n')).toEqual({ args: ['next'] });
    expect(parseAllFoursCommand('next')).toEqual({ args: ['next'] });
    expect(parseAllFoursCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseAllFoursCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses hint, log, reset', () => {
    expect(parseAllFoursCommand('h')).toEqual({ args: ['hint'] });
    expect(parseAllFoursCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseAllFoursCommand('log')).toEqual({ args: ['log'] });
    expect(parseAllFoursCommand('r')).toEqual({ args: ['reset'] });
    expect(parseAllFoursCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns suggestion error for near-miss command', () => {
    const result = parseAllFoursCommand('stnd');
    expect('error' in result).toBe(true);
  });

  it('returns error for unknown command', () => {
    const result = parseAllFoursCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('exposes help text', () => {
    expect(ALLFOURS_HELP.length).toBeGreaterThan(0);
  });
});
