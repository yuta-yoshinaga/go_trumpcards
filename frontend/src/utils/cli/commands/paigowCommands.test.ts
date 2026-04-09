import { describe, expect, it } from 'vitest';
import { parsePaigowCommand } from './paigowCommands';

describe('parsePaigowCommand', () => {
  it('parses bet with amount', () => {
    expect(parsePaigowCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parsePaigowCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('returns error for bet without amount', () => {
    const result = parsePaigowCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses set with two indices', () => {
    expect(parsePaigowCommand('s 0 1')).toEqual({ args: ['set', undefined, 0, 1] });
    expect(parsePaigowCommand('set 3 5')).toEqual({ args: ['set', undefined, 3, 5] });
  });

  it('returns error for set without indices', () => {
    expect('error' in parsePaigowCommand('s')).toBe(true);
    expect('error' in parsePaigowCommand('s 0')).toBe(true);
  });

  it('parses log', () => {
    expect(parsePaigowCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parsePaigowCommand('r')).toEqual({ args: ['reset'] });
    expect(parsePaigowCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error with suggestion for mistyped command', () => {
    const result = parsePaigowCommand('bett');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toContain('Did you mean');
    }
  });

  it('returns error for unknown command', () => {
    const result = parsePaigowCommand('xyz');
    expect('error' in result).toBe(true);
    if ('error' in result) {
      expect(result.error).toContain('Unknown command');
    }
  });
});
