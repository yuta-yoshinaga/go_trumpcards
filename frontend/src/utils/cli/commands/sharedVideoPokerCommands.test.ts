import { describe, expect, it } from 'vitest';
import { parseVideoPokerCommand } from './sharedVideoPokerCommands';

describe('parseVideoPokerCommand', () => {
  it('parses bet with amount', () => {
    expect(parseVideoPokerCommand('b 100')).toEqual({ command: 'bet', amount: 100 });
    expect(parseVideoPokerCommand('bet 50')).toEqual({ command: 'bet', amount: 50 });
  });

  it('returns error for bet without amount', () => {
    const result = parseVideoPokerCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses hold with indices', () => {
    expect(parseVideoPokerCommand('hold 0 2 4')).toEqual({ command: 'hold', indices: [0, 2, 4] });
  });

  it('parses hold without indices as empty', () => {
    expect(parseVideoPokerCommand('hold')).toEqual({ command: 'hold', indices: [] });
  });

  it('parses reset', () => {
    expect(parseVideoPokerCommand('r')).toEqual({ command: 'reset' });
    expect(parseVideoPokerCommand('reset')).toEqual({ command: 'reset' });
  });

  it('returns error for unknown command', () => {
    const result = parseVideoPokerCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
