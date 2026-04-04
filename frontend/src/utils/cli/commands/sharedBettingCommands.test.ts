import { describe, expect, it } from 'vitest';
import { parseBettingCommand } from './sharedBettingCommands';

describe('parseBettingCommand', () => {
  it('parses fold', () => {
    expect(parseBettingCommand('f')).toEqual({ command: 'fold' });
    expect(parseBettingCommand('fold')).toEqual({ command: 'fold' });
  });

  it('parses check', () => {
    expect(parseBettingCommand('ck')).toEqual({ command: 'check' });
    expect(parseBettingCommand('check')).toEqual({ command: 'check' });
  });

  it('parses call', () => {
    expect(parseBettingCommand('c')).toEqual({ command: 'call' });
    expect(parseBettingCommand('call')).toEqual({ command: 'call' });
  });

  it('parses allin', () => {
    expect(parseBettingCommand('a')).toEqual({ command: 'allin' });
    expect(parseBettingCommand('allin')).toEqual({ command: 'allin' });
  });

  it('parses bet with amount', () => {
    expect(parseBettingCommand('b 100')).toEqual({ command: 'bet', amount: 100 });
    expect(parseBettingCommand('bet 50')).toEqual({ command: 'bet', amount: 50 });
  });

  it('returns error for bet without amount', () => {
    const result = parseBettingCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses raise with amount', () => {
    expect(parseBettingCommand('ra 200')).toEqual({ command: 'raise', amount: 200 });
    expect(parseBettingCommand('raise 300')).toEqual({ command: 'raise', amount: 300 });
  });

  it('returns error for raise without amount', () => {
    const result = parseBettingCommand('ra');
    expect('error' in result).toBe(true);
  });

  it('parses reset', () => {
    expect(parseBettingCommand('r')).toEqual({ command: 'reset' });
    expect(parseBettingCommand('reset')).toEqual({ command: 'reset' });
  });

  it('returns error for unknown command', () => {
    const result = parseBettingCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('delegates to extra parser', () => {
    const result = parseBettingCommand('log', ['log'], (cmd) => {
      if (cmd === 'log') return { command: 'log' };
      return null;
    });
    expect(result).toEqual({ command: 'log' });
  });

  it('extra parser error is returned', () => {
    const result = parseBettingCommand('log', ['log'], () => {
      return { error: 'custom error' };
    });
    expect('error' in result).toBe(true);
  });
});
