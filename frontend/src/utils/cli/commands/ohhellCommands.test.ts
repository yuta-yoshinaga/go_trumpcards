import { describe, expect, it } from 'vitest';
import { parseOhhellCommand } from './ohhellCommands';

describe('parseOhhellCommand', () => {
  it('parses bid with number', () => {
    expect(parseOhhellCommand('bid 3')).toEqual({ args: ['bid', 3] });
  });

  it('returns error for bid without number', () => {
    const result = parseOhhellCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('parses play with index', () => {
    expect(parseOhhellCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
    expect(parseOhhellCommand('play 5')).toEqual({ args: ['play', undefined, 5] });
  });

  it('returns error for play without index', () => {
    const result = parseOhhellCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next', () => {
    expect(parseOhhellCommand('n')).toEqual({ args: ['next', undefined, undefined] });
  });

  it('parses nextround', () => {
    expect(parseOhhellCommand('nr')).toEqual({ args: ['nextround', undefined, undefined] });
  });

  it('parses hint', () => {
    expect(parseOhhellCommand('h')).toEqual({ args: ['hint', undefined, undefined] });
  });

  it('parses reset', () => {
    expect(parseOhhellCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseOhhellCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
