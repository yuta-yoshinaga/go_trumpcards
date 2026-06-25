import { describe, expect, it } from 'vitest';
import { HIGHCARDFLUSH_HELP, parseHighcardflushCommand } from './highcardflushCommands';

describe('parseHighcardflushCommand', () => {
  it('parses bet with ante only', () => {
    expect(parseHighcardflushCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseHighcardflushCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('parses bet with ante and flush bonus side bet', () => {
    expect(parseHighcardflushCommand('b 100 10')).toEqual({ args: ['bet', 100, 10] });
  });

  it('parses bet with ante, flush bonus, and straight flush side bets', () => {
    expect(parseHighcardflushCommand('b 100 10 5')).toEqual({ args: ['bet', 100, 10, 5] });
  });

  it('returns error for bet without amount', () => {
    expect('error' in parseHighcardflushCommand('b')).toBe(true);
    expect('error' in parseHighcardflushCommand('bet')).toBe(true);
  });

  it('returns error for non-numeric ante', () => {
    expect('error' in parseHighcardflushCommand('b abc')).toBe(true);
  });

  it('returns error for non-numeric side bet', () => {
    expect('error' in parseHighcardflushCommand('b 100 xyz')).toBe(true);
    expect('error' in parseHighcardflushCommand('b 100 10 xyz')).toBe(true);
  });

  it('parses numeric raise shortcuts 1/2/3', () => {
    expect(parseHighcardflushCommand('1')).toEqual({ args: ['raise', undefined, undefined, undefined, 1] });
    expect(parseHighcardflushCommand('2')).toEqual({ args: ['raise', undefined, undefined, undefined, 2] });
    expect(parseHighcardflushCommand('3')).toEqual({ args: ['raise', undefined, undefined, undefined, 3] });
  });

  it('parses raise <multiplier>', () => {
    expect(parseHighcardflushCommand('ra 2')).toEqual({ args: ['raise', undefined, undefined, undefined, 2] });
    expect(parseHighcardflushCommand('raise 3')).toEqual({ args: ['raise', undefined, undefined, undefined, 3] });
  });

  it('returns error for out-of-range raise multiplier', () => {
    expect('error' in parseHighcardflushCommand('raise 4')).toBe(true);
    expect('error' in parseHighcardflushCommand('raise 0')).toBe(true);
  });

  it('returns error for raise without a multiplier', () => {
    expect('error' in parseHighcardflushCommand('raise')).toBe(true);
  });

  it('parses fold', () => {
    expect(parseHighcardflushCommand('f')).toEqual({ args: ['fold'] });
    expect(parseHighcardflushCommand('fold')).toEqual({ args: ['fold'] });
  });

  it('parses log', () => {
    expect(parseHighcardflushCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseHighcardflushCommand('r')).toEqual({ args: ['reset'] });
    expect(parseHighcardflushCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const result = parseHighcardflushCommand('fol');
    expect('error' in result && result.error).toContain('fold');
  });

  it('returns error for unknown command', () => {
    expect('error' in parseHighcardflushCommand('xyz')).toBe(true);
  });

  it('exposes help text', () => {
    expect(HIGHCARDFLUSH_HELP.length).toBeGreaterThan(0);
    expect(HIGHCARDFLUSH_HELP.some((line) => line.includes('fold'))).toBe(true);
  });
});
