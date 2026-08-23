import { describe, expect, it } from 'vitest';
import { parseMadrassoCommand } from './madrassoCommands';

describe('parseMadrassoCommand', () => {
  it('parses play', () => {
    expect(parseMadrassoCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
  });

  it('parses play long form', () => {
    expect(parseMadrassoCommand('play 0')).toEqual({ args: ['play', undefined, 0] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseMadrassoCommand('p')).toBe(true);
  });

  it('parses next', () => {
    expect(parseMadrassoCommand('n')).toEqual({ args: ['next'] });
  });

  it('parses nextround', () => {
    expect(parseMadrassoCommand('nr')).toEqual({ args: ['nextround'] });
  });

  it('parses hint', () => {
    expect(parseMadrassoCommand('h')).toEqual({ args: ['hint'] });
  });

  it('parses reset', () => {
    expect(parseMadrassoCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseMadrassoCommand('paly 1');
    expect('error' in result).toBe(true);
  });

  it('returns error for unknown command', () => {
    expect('error' in parseMadrassoCommand('xyz')).toBe(true);
  });
});
