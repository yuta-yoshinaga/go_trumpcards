import { describe, expect, it } from 'vitest';
import { parseSevensCommand } from './sevensCommands';

describe('parseSevensCommand', () => {
  it('parses play with index', () => {
    expect(parseSevensCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parseSevensCommand('play 5')).toEqual({ args: ['play', 5] });
  });

  it('returns error for play without index', () => {
    const result = parseSevensCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses pass', () => {
    expect(parseSevensCommand('pass')).toEqual({ args: ['play', -1] });
  });

  it('parses joker with suit and value', () => {
    expect(parseSevensCommand('j spade 6')).toEqual({ args: ['joker', -1, 1, 6] });
    expect(parseSevensCommand('joker heart 3')).toEqual({ args: ['joker', -1, 3, 3] });
    expect(parseSevensCommand('j d 10')).toEqual({ args: ['joker', -1, 4, 10] });
  });

  it('returns error for joker without enough args', () => {
    const result = parseSevensCommand('j');
    expect('error' in result).toBe(true);
    const result2 = parseSevensCommand('j spade');
    expect('error' in result2).toBe(true);
  });

  // Regression for #2152: `club` (singular) must be accepted, matching Mighty.
  it('accepts club, clubs, and clover for the joker suit', () => {
    expect(parseSevensCommand('j club 6')).toEqual({ args: ['joker', -1, 2, 6] });
    expect(parseSevensCommand('j clubs 6')).toEqual({ args: ['joker', -1, 2, 6] });
    expect(parseSevensCommand('j clover 6')).toEqual({ args: ['joker', -1, 2, 6] });
  });

  it('returns error for joker with invalid suit', () => {
    const result = parseSevensCommand('j invalid 5');
    expect('error' in result).toBe(true);
  });

  it('parses reset', () => {
    expect(parseSevensCommand('r')).toEqual({ args: ['reset'] });
    expect(parseSevensCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseSevensCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
