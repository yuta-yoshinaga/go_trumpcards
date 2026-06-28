import { describe, expect, it } from 'vitest';
import { parsePrsiCommand } from './prsiCommands';

describe('parsePrsiCommand', () => {
  it('parses play with index', () => {
    expect(parsePrsiCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parsePrsiCommand('play 5')).toEqual({ args: ['play', 5] });
  });

  it('returns error for play without index', () => {
    const result = parsePrsiCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses draw', () => {
    expect(parsePrsiCommand('d')).toEqual({ args: ['draw'] });
    expect(parsePrsiCommand('draw')).toEqual({ args: ['draw'] });
  });

  it('parses reset', () => {
    expect(parsePrsiCommand('r')).toEqual({ args: ['reset'] });
    expect(parsePrsiCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('parses log', () => {
    expect(parsePrsiCommand('l')).toEqual({ args: ['log'] });
    expect(parsePrsiCommand('log')).toEqual({ args: ['log'] });
  });

  it('returns error for unknown command', () => {
    const result = parsePrsiCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('suggests a close command for a typo', () => {
    const result = parsePrsiCommand('ply 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('play');
  });
});
