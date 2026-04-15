import { describe, expect, it } from 'vitest';
import { POKERSQUARES_HELP, parsePokerSquaresCommand } from './pokersquaresCommands';

describe('parsePokerSquaresCommand', () => {
  it('parses place with row/col (p alias)', () => {
    expect(parsePokerSquaresCommand('p 2 3')).toEqual({ args: ['place', 2, 3] });
  });

  it('parses place with row/col (place alias)', () => {
    expect(parsePokerSquaresCommand('place 0 0')).toEqual({ args: ['place', 0, 0] });
  });

  it('returns error for place with missing args', () => {
    const result = parsePokerSquaresCommand('p 2');
    expect('error' in result).toBe(true);
  });

  it('returns error for place with non-integer args', () => {
    const result = parsePokerSquaresCommand('p a b');
    expect('error' in result).toBe(true);
  });

  it('returns error for out-of-range row', () => {
    const result = parsePokerSquaresCommand('p 5 0');
    expect('error' in result).toBe(true);
  });

  it('returns error for out-of-range col', () => {
    const result = parsePokerSquaresCommand('p 0 -1');
    expect('error' in result).toBe(true);
  });

  it('parses undo (u alias)', () => {
    expect(parsePokerSquaresCommand('u')).toEqual({ args: ['undo'] });
  });

  it('parses undo (undo alias)', () => {
    expect(parsePokerSquaresCommand('undo')).toEqual({ args: ['undo'] });
  });

  it('parses giveup (g alias)', () => {
    expect(parsePokerSquaresCommand('g')).toEqual({ args: ['giveup'] });
  });

  it('parses giveup (giveup alias)', () => {
    expect(parsePokerSquaresCommand('giveup')).toEqual({ args: ['giveup'] });
  });

  it('parses log', () => {
    expect(parsePokerSquaresCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (r alias)', () => {
    expect(parsePokerSquaresCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses reset (reset alias)', () => {
    expect(parsePokerSquaresCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error with suggestion for close typo', () => {
    const result = parsePokerSquaresCommand('rese');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toContain('Did you mean');
    }
  });

  it('returns error without suggestion for unknown command', () => {
    const result = parsePokerSquaresCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).not.toContain('Did you mean');
    }
  });
});

describe('POKERSQUARES_HELP', () => {
  it('is a non-empty array of help strings', () => {
    expect(Array.isArray(POKERSQUARES_HELP)).toBe(true);
    expect(POKERSQUARES_HELP.length).toBeGreaterThan(0);
    for (const line of POKERSQUARES_HELP) {
      expect(typeof line).toBe('string');
    }
  });
});
