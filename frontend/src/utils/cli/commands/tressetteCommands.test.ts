import { describe, expect, it } from 'vitest';
import { parseTressetteCommand } from './tressetteCommands';

describe('parseTressetteCommand', () => {
  it('parses play', () => {
    expect(parseTressetteCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
  });

  it('parses play long form', () => {
    expect(parseTressetteCommand('play 0')).toEqual({ args: ['play', undefined, 0] });
  });

  it('returns error for play without index', () => {
    expect('error' in parseTressetteCommand('p')).toBe(true);
  });

  it('parses next', () => {
    expect(parseTressetteCommand('n')).toEqual({ args: ['next'] });
  });

  it('parses nextround', () => {
    expect(parseTressetteCommand('nr')).toEqual({ args: ['nextround'] });
  });

  it('parses hint', () => {
    expect(parseTressetteCommand('h')).toEqual({ args: ['hint'] });
  });

  it('parses reset', () => {
    expect(parseTressetteCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests for a near-miss command', () => {
    const result = parseTressetteCommand('paly 1');
    expect('error' in result).toBe(true);
  });

  it('returns error for unknown command', () => {
    expect('error' in parseTressetteCommand('xyz')).toBe(true);
  });
});
