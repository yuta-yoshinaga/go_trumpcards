import { describe, expect, it } from 'vitest';
import { parseReddogCommand, REDDOG_HELP } from './reddogCommands';

describe('parseReddogCommand', () => {
  it('parses bet with amount (b alias)', () => {
    expect(parseReddogCommand('b 100')).toEqual({ args: ['bet', 100] });
  });

  it('parses bet with amount (bet alias)', () => {
    expect(parseReddogCommand('bet 200')).toEqual({ args: ['bet', 200] });
  });

  it('returns error for bet without amount', () => {
    const result = parseReddogCommand('b');
    expect('error' in result).toBe(true);
  });

  it('returns error for bet with non-integer amount', () => {
    const result = parseReddogCommand('b abc');
    expect('error' in result).toBe(true);
  });

  it('parses raise with amount', () => {
    expect(parseReddogCommand('raise 50')).toEqual({ args: ['raise', 50] });
  });

  it('returns error for raise without amount', () => {
    const result = parseReddogCommand('raise');
    expect('error' in result).toBe(true);
  });

  it('parses stay (s alias)', () => {
    expect(parseReddogCommand('s')).toEqual({ args: ['stay'] });
  });

  it('parses stay (stay alias)', () => {
    expect(parseReddogCommand('stay')).toEqual({ args: ['stay'] });
  });

  it('parses log', () => {
    expect(parseReddogCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (r alias)', () => {
    expect(parseReddogCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses reset (reset alias)', () => {
    expect(parseReddogCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error with suggestion for close typo', () => {
    const result = parseReddogCommand('rese');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toContain('Did you mean');
    }
  });

  it('returns error without suggestion for unknown command', () => {
    const result = parseReddogCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).not.toContain('Did you mean');
    }
  });
});

describe('REDDOG_HELP', () => {
  it('is a non-empty array of help strings', () => {
    expect(Array.isArray(REDDOG_HELP)).toBe(true);
    expect(REDDOG_HELP.length).toBeGreaterThan(0);
    for (const line of REDDOG_HELP) {
      expect(typeof line).toBe('string');
    }
  });
});
