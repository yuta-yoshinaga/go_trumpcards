import { describe, expect, it } from 'vitest';
import { parseShortdeckCommand } from './shortdeckCommands';

describe('parseShortdeckCommand', () => {
  it('parses fold', () => {
    expect(parseShortdeckCommand('f')).toEqual({ args: ['fold', undefined] });
  });

  it('parses check', () => {
    expect(parseShortdeckCommand('ck')).toEqual({ args: ['check', undefined] });
  });

  it('parses call', () => {
    expect(parseShortdeckCommand('c')).toEqual({ args: ['call', undefined] });
  });

  it('parses bet with amount', () => {
    expect(parseShortdeckCommand('b 100')).toEqual({ args: ['bet', 100] });
  });

  it('returns error for bet without amount', () => {
    const result = parseShortdeckCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses raise with amount', () => {
    expect(parseShortdeckCommand('ra 200')).toEqual({ args: ['raise', 200] });
  });

  it('parses allin', () => {
    expect(parseShortdeckCommand('a')).toEqual({ args: ['allin', undefined] });
  });

  it('parses reset', () => {
    expect(parseShortdeckCommand('r')).toEqual({ args: ['reset', undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseShortdeckCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
