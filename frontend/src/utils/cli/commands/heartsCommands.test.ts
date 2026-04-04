import { describe, expect, it } from 'vitest';
import { parseHeartsCommand } from './heartsCommands';

describe('parseHeartsCommand', () => {
  it('parses pass with 3 indices', () => {
    expect(parseHeartsCommand('pass 0 3 5')).toEqual({ args: ['pass', [0, 3, 5]] });
  });

  it('returns error for pass with wrong count', () => {
    const result = parseHeartsCommand('pass 0 1');
    expect('error' in result).toBe(true);
  });

  it('parses play', () => {
    expect(parseHeartsCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
  });

  it('returns error for play without index', () => {
    const result = parseHeartsCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next', () => {
    expect(parseHeartsCommand('n')).toEqual({ args: ['next'] });
  });

  it('parses nextround', () => {
    expect(parseHeartsCommand('nr')).toEqual({ args: ['nextround'] });
  });

  it('parses hint', () => {
    expect(parseHeartsCommand('h')).toEqual({ args: ['hint'] });
  });

  it('parses reset', () => {
    expect(parseHeartsCommand('r')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseHeartsCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
