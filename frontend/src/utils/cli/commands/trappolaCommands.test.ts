import { describe, expect, it } from 'vitest';
import { parseTrappolaCommand } from './trappolaCommands';

describe('parseTrappolaCommand', () => {
  it('parses play', () => {
    expect(parseTrappolaCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
  });

  it('parses play long form', () => {
    expect(parseTrappolaCommand('play 0')).toEqual({ args: ['play', undefined, 0] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseTrappolaCommand('p')).toBe(true);
  });

  it('parses next', () => {
    expect(parseTrappolaCommand('n')).toEqual({ args: ['next'] });
  });

  it('parses nextround', () => {
    expect(parseTrappolaCommand('nr')).toEqual({ args: ['nextround'] });
  });

  it('parses hint', () => {
    expect(parseTrappolaCommand('h')).toEqual({ args: ['hint'] });
  });

  it('parses reset', () => {
    expect(parseTrappolaCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseTrappolaCommand('paly 1');
    expect('error' in result).toBe(true);
  });

  it('returns error for unknown command', () => {
    expect('error' in parseTrappolaCommand('xyz')).toBe(true);
  });
});
