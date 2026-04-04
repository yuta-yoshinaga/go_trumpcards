import { describe, expect, it } from 'vitest';
import { parseOmahaCommand } from './omahaCommands';

describe('parseOmahaCommand', () => {
  it('parses fold', () => {
    expect(parseOmahaCommand('f')).toEqual({ args: ['fold', undefined] });
  });

  it('parses check', () => {
    expect(parseOmahaCommand('ck')).toEqual({ args: ['check', undefined] });
  });

  it('parses call', () => {
    expect(parseOmahaCommand('c')).toEqual({ args: ['call', undefined] });
  });

  it('parses bet with amount', () => {
    expect(parseOmahaCommand('b 100')).toEqual({ args: ['bet', 100] });
  });

  it('returns error for bet without amount', () => {
    const result = parseOmahaCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses raise with amount', () => {
    expect(parseOmahaCommand('ra 200')).toEqual({ args: ['raise', 200] });
  });

  it('parses allin', () => {
    expect(parseOmahaCommand('a')).toEqual({ args: ['allin', undefined] });
  });

  it('parses reset', () => {
    expect(parseOmahaCommand('r')).toEqual({ args: ['reset', undefined] });
  });

  it('returns error for unknown command', () => {
    const result = parseOmahaCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
