import { describe, expect, it } from 'vitest';
import { parseSpadesCommand } from './spadesCommands';

describe('parseSpadesCommand', () => {
  it('parses bid with number', () => {
    expect(parseSpadesCommand('bid 3')).toEqual({ args: ['bid', 3] });
  });

  it('returns error for bid without number', () => {
    const result = parseSpadesCommand('bid');
    expect('error' in result).toBe(true);
  });

  it('parses play with index', () => {
    expect(parseSpadesCommand('p 2')).toEqual({ args: ['play', undefined, 2] });
    expect(parseSpadesCommand('play 5')).toEqual({ args: ['play', undefined, 5] });
  });

  it('returns error for play without index', () => {
    const result = parseSpadesCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next', () => {
    expect(parseSpadesCommand('n')).toEqual({ args: ['next', undefined, undefined] });
  });

  it('parses nextround', () => {
    expect(parseSpadesCommand('nr')).toEqual({ args: ['nextround', undefined, undefined] });
  });

  it('parses hint', () => {
    expect(parseSpadesCommand('h')).toEqual({ args: ['hint', undefined, undefined] });
  });

  it('parses reset', () => {
    expect(parseSpadesCommand('r')).toEqual({ args: ['reset', undefined, undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseSpadesCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
