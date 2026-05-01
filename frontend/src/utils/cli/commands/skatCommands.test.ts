import { describe, expect, it } from 'vitest';
import { parseSkatCommand } from './skatCommands';

describe('parseSkatCommand', () => {
  it('parses bid commands', () => {
    expect(parseSkatCommand('b')).toEqual({ args: ['bid', { accept: true }] });
    expect(parseSkatCommand('bid')).toEqual({ args: ['bid', { accept: true }] });
    expect(parseSkatCommand('pa')).toEqual({ args: ['bid', { accept: false }] });
    expect(parseSkatCommand('pass')).toEqual({ args: ['bid', { accept: false }] });
  });

  it('parses skat pickup', () => {
    expect(parseSkatCommand('ps')).toEqual({ args: ['pickskat', { pickup: true }] });
    expect(parseSkatCommand('sk')).toEqual({ args: ['pickskat', { pickup: false }] });
  });

  it('parses discard', () => {
    expect(parseSkatCommand('d 0 3')).toEqual({ args: ['discard', { discardA: 0, discardB: 3 }] });
  });

  it('rejects discard without two indices', () => {
    expect('error' in parseSkatCommand('d')).toBe(true);
    expect('error' in parseSkatCommand('d 0')).toBe(true);
  });

  it('parses game declaration', () => {
    expect(parseSkatCommand('g 1 2')).toEqual({
      args: ['game', { gameType: 1, trumpSuit: 2 }],
    });
    expect(parseSkatCommand('g 2')).toEqual({
      args: ['game', { gameType: 2, trumpSuit: undefined }],
    });
  });

  it('parses play', () => {
    expect(parseSkatCommand('p 5')).toEqual({ args: ['play', { cardIndex: 5 }] });
  });

  it('rejects play without index', () => {
    expect('error' in parseSkatCommand('p')).toBe(true);
  });

  it('parses control commands', () => {
    expect(parseSkatCommand('n')).toEqual({ args: ['next'] });
    expect(parseSkatCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseSkatCommand('h')).toEqual({ args: ['hint'] });
    expect(parseSkatCommand('log')).toEqual({ args: ['log'] });
    expect(parseSkatCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests near-match for typos', () => {
    const result = parseSkatCommand('passs');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('Did you mean');
  });

  it('returns error for unknown command', () => {
    const result = parseSkatCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
