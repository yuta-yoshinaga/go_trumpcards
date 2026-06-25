import { describe, expect, it } from 'vitest';
import { OPENFACECHINESE_HELP, parseOpenfacechineseCommand } from './openfacechineseCommands';

describe('parseOpenfacechineseCommand', () => {
  it('parses place with a row name', () => {
    expect(parseOpenfacechineseCommand('place front')).toEqual({ args: ['place', { row: 0 }] });
    expect(parseOpenfacechineseCommand('place middle')).toEqual({ args: ['place', { row: 1 }] });
    expect(parseOpenfacechineseCommand('place back')).toEqual({ args: ['place', { row: 2 }] });
  });

  it('parses place with row initials and indices', () => {
    expect(parseOpenfacechineseCommand('p f')).toEqual({ args: ['place', { row: 0 }] });
    expect(parseOpenfacechineseCommand('p m')).toEqual({ args: ['place', { row: 1 }] });
    expect(parseOpenfacechineseCommand('p b')).toEqual({ args: ['place', { row: 2 }] });
    expect(parseOpenfacechineseCommand('p 2')).toEqual({ args: ['place', { row: 2 }] });
  });

  it('returns error for place without a valid row', () => {
    expect('error' in parseOpenfacechineseCommand('place')).toBe(true);
    expect('error' in parseOpenfacechineseCommand('place sideways')).toBe(true);
  });

  it('parses next (and aliases)', () => {
    expect(parseOpenfacechineseCommand('n')).toEqual({ args: ['nextround'] });
    expect(parseOpenfacechineseCommand('next')).toEqual({ args: ['nextround'] });
    expect(parseOpenfacechineseCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseOpenfacechineseCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseOpenfacechineseCommand('r')).toEqual({ args: ['reset'] });
    expect(parseOpenfacechineseCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command for a typo', () => {
    const result = parseOpenfacechineseCommand('rese');
    expect('error' in result && result.error).toContain('reset');
  });

  it('returns error for unknown command', () => {
    expect('error' in parseOpenfacechineseCommand('xyz')).toBe(true);
  });

  it('exposes help text', () => {
    expect(OPENFACECHINESE_HELP.length).toBeGreaterThan(0);
    expect(OPENFACECHINESE_HELP.some((line) => line.includes('place'))).toBe(true);
  });
});
