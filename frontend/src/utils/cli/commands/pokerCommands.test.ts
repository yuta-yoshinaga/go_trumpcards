import { describe, expect, it } from 'vitest';
import { parsePokerCommand } from './pokerCommands';

describe('parsePokerCommand', () => {
  it('parses exchange with indices', () => {
    expect(parsePokerCommand('e 0 2')).toEqual({ args: ['exchange', [0, 2]] });
  });

  it('parses stand', () => {
    expect(parsePokerCommand('s')).toEqual({ args: ['stand'] });
  });

  it('parses bet with amount', () => {
    expect(parsePokerCommand('b 100')).toEqual({ args: ['bet', undefined, 100] });
  });

  it('parses call', () => {
    expect(parsePokerCommand('c')).toEqual({ args: ['call'] });
  });

  it('parses raise', () => {
    expect(parsePokerCommand('ra 200')).toEqual({ args: ['raise', undefined, 200] });
  });

  it('parses fold', () => {
    expect(parsePokerCommand('f')).toEqual({ args: ['fold'] });
  });

  it('parses check', () => {
    expect(parsePokerCommand('ck')).toEqual({ args: ['check'] });
  });

  it('parses allin', () => {
    expect(parsePokerCommand('a')).toEqual({ args: ['allin'] });
  });

  it('parses odds', () => {
    expect(parsePokerCommand('o 0 3')).toEqual({ args: ['odds', [0, 3]] });
  });

  it('parses reset', () => {
    expect(parsePokerCommand('r')).toEqual({ args: ['reset'] });
  });

  it('returns error for bet without amount', () => {
    const result = parsePokerCommand('b');
    expect('error' in result).toBe(true);
  });

  it('returns error for unknown command', () => {
    const result = parsePokerCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
