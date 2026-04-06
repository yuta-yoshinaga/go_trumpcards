import { describe, expect, it } from 'vitest';
import { parseSpeedCommand } from './speedCommands';

describe('parseSpeedCommand', () => {
  it('parses play with card and pile indices', () => {
    expect(parseSpeedCommand('p 0 1')).toEqual({ args: ['play', 0, 1] });
    expect(parseSpeedCommand('play 3 0')).toEqual({ args: ['play', 3, 0] });
  });

  it('returns error for play without enough args', () => {
    const result = parseSpeedCommand('p');
    expect('error' in result).toBe(true);
    const result2 = parseSpeedCommand('p 0');
    expect('error' in result2).toBe(true);
  });

  it('parses flip', () => {
    expect(parseSpeedCommand('fl')).toEqual({ args: ['flip'] });
    expect(parseSpeedCommand('flip')).toEqual({ args: ['flip'] });
  });

  it('parses hint', () => {
    expect(parseSpeedCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSpeedCommand('hint')).toEqual({ args: ['hint'] });
  });

  it('parses log', () => {
    expect(parseSpeedCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseSpeedCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSpeedCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseSpeedCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
