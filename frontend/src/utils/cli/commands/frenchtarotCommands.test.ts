import { describe, expect, it } from 'vitest';
import { FRENCH_TAROT_HELP, parseFrenchTarotCommand } from './frenchtarotCommands';

describe('parseFrenchTarotCommand', () => {
  it('parses bid contracts (full names)', () => {
    expect(parseFrenchTarotCommand('bid petite')).toEqual({ args: ['bid', { bid: 'petite' }] });
    expect(parseFrenchTarotCommand('bid garde')).toEqual({ args: ['bid', { bid: 'garde' }] });
    expect(parseFrenchTarotCommand('bid gardesans')).toEqual({ args: ['bid', { bid: 'gardesans' }] });
    expect(parseFrenchTarotCommand('bid gardecontre')).toEqual({ args: ['bid', { bid: 'gardecontre' }] });
  });

  it('parses bid shorthands and hyphenated forms', () => {
    expect(parseFrenchTarotCommand('b p')).toEqual({ args: ['bid', { bid: 'petite' }] });
    expect(parseFrenchTarotCommand('b g')).toEqual({ args: ['bid', { bid: 'garde' }] });
    expect(parseFrenchTarotCommand('b gs')).toEqual({ args: ['bid', { bid: 'gardesans' }] });
    expect(parseFrenchTarotCommand('b gc')).toEqual({ args: ['bid', { bid: 'gardecontre' }] });
    expect(parseFrenchTarotCommand('bid garde-sans')).toEqual({ args: ['bid', { bid: 'gardesans' }] });
    expect(parseFrenchTarotCommand('bid garde-contre')).toEqual({ args: ['bid', { bid: 'gardecontre' }] });
  });

  it('rejects an unknown or missing bid', () => {
    expect(parseFrenchTarotCommand('bid')).toHaveProperty('error');
    expect(parseFrenchTarotCommand('bid grande')).toHaveProperty('error');
  });

  it('parses pass', () => {
    expect(parseFrenchTarotCommand('pass')).toEqual({ args: ['pass'] });
  });

  it('parses a 6-card discard (écart)', () => {
    expect(parseFrenchTarotCommand('discard 0 1 2 3 4 5')).toEqual({
      args: ['discard', { cardIndices: [0, 1, 2, 3, 4, 5] }],
    });
    expect(parseFrenchTarotCommand('d 10 11 12 13 14 15')).toEqual({
      args: ['discard', { cardIndices: [10, 11, 12, 13, 14, 15] }],
    });
  });

  it('rejects a discard with fewer than 6 indices', () => {
    expect(parseFrenchTarotCommand('discard 0 1 2')).toHaveProperty('error');
    expect(parseFrenchTarotCommand('discard 0 1 2 3 4 x')).toHaveProperty('error');
  });

  it('parses play with alias', () => {
    expect(parseFrenchTarotCommand('p 3')).toEqual({ args: ['play', { cardIndex: 3 }] });
    expect(parseFrenchTarotCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('rejects play without a valid index', () => {
    expect(parseFrenchTarotCommand('p')).toHaveProperty('error');
    expect(parseFrenchTarotCommand('play x')).toHaveProperty('error');
  });

  it('parses next / nextround / hint / reset with aliases', () => {
    expect(parseFrenchTarotCommand('n')).toEqual({ args: ['next'] });
    expect(parseFrenchTarotCommand('next')).toEqual({ args: ['next'] });
    expect(parseFrenchTarotCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseFrenchTarotCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseFrenchTarotCommand('h')).toEqual({ args: ['hint'] });
    expect(parseFrenchTarotCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseFrenchTarotCommand('r')).toEqual({ args: ['reset'] });
    expect(parseFrenchTarotCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const res = parseFrenchTarotCommand('discrd 0 1 2 3 4 5');
    expect(res).toHaveProperty('error');
    if ('error' in res) expect(res.error).toContain('discard');
  });

  it('reports an unknown command with no close match', () => {
    const res = parseFrenchTarotCommand('zzz');
    expect(res).toHaveProperty('error');
    if ('error' in res) expect(res.error).toContain('Unknown command');
  });

  it('exposes help text', () => {
    expect(FRENCH_TAROT_HELP.length).toBeGreaterThan(0);
    expect(FRENCH_TAROT_HELP.some((l) => l.includes('gardecontre'))).toBe(true);
  });
});
