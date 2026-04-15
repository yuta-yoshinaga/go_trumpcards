import { describe, expect, it } from 'vitest';
import { LETITRIDE_HELP, parseLetitrideCommand } from './letitrideCommands';

describe('parseLetitrideCommand', () => {
  it('parses bet with amount (b alias)', () => {
    expect(parseLetitrideCommand('b 100')).toEqual({ args: ['bet', 100] });
  });

  it('parses bet with amount (bet alias)', () => {
    expect(parseLetitrideCommand('bet 200')).toEqual({ args: ['bet', 200] });
  });

  it('returns error for bet without amount', () => {
    const result = parseLetitrideCommand('b');
    expect('error' in result).toBe(true);
  });

  it('returns error for bet with non-integer amount', () => {
    const result = parseLetitrideCommand('b abc');
    expect('error' in result).toBe(true);
  });

  it('parses pull (p alias)', () => {
    expect(parseLetitrideCommand('p')).toEqual({ args: ['pull'] });
  });

  it('parses pull (pull alias)', () => {
    expect(parseLetitrideCommand('pull')).toEqual({ args: ['pull'] });
  });

  it('parses letitride (l alias)', () => {
    expect(parseLetitrideCommand('l')).toEqual({ args: ['letitride'] });
  });

  it('parses letitride (letitride alias)', () => {
    expect(parseLetitrideCommand('letitride')).toEqual({ args: ['letitride'] });
  });

  it('parses log', () => {
    expect(parseLetitrideCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset (r alias)', () => {
    expect(parseLetitrideCommand('r')).toEqual({ args: ['reset'] });
  });

  it('parses reset (reset alias)', () => {
    expect(parseLetitrideCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error with suggestion for close typo', () => {
    const result = parseLetitrideCommand('rese');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toContain('Did you mean');
    }
  });

  it('returns error without suggestion for unknown command', () => {
    const result = parseLetitrideCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).not.toContain('Did you mean');
    }
  });
});

describe('LETITRIDE_HELP', () => {
  it('is a non-empty array of help strings', () => {
    expect(Array.isArray(LETITRIDE_HELP)).toBe(true);
    expect(LETITRIDE_HELP.length).toBeGreaterThan(0);
    for (const line of LETITRIDE_HELP) {
      expect(typeof line).toBe('string');
    }
  });
});
