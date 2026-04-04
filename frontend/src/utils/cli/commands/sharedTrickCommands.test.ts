import { describe, expect, it } from 'vitest';
import { parseTrickCommand } from './sharedTrickCommands';

describe('parseTrickCommand', () => {
  it('parses play with index', () => {
    expect(parseTrickCommand('p 2')).toEqual({ command: 'play', cardIndex: 2 });
    expect(parseTrickCommand('play 5')).toEqual({ command: 'play', cardIndex: 5 });
  });

  it('returns error for play without index', () => {
    const result = parseTrickCommand('p');
    expect('error' in result).toBe(true);
  });

  it('parses next', () => {
    expect(parseTrickCommand('n')).toEqual({ command: 'next' });
    expect(parseTrickCommand('next')).toEqual({ command: 'next' });
  });

  it('parses nextround', () => {
    expect(parseTrickCommand('nr')).toEqual({ command: 'nextround' });
    expect(parseTrickCommand('nextround')).toEqual({ command: 'nextround' });
  });

  it('parses hint', () => {
    expect(parseTrickCommand('h')).toEqual({ command: 'hint' });
    expect(parseTrickCommand('hint')).toEqual({ command: 'hint' });
  });

  it('parses reset', () => {
    expect(parseTrickCommand('r')).toEqual({ command: 'reset' });
    expect(parseTrickCommand('reset')).toEqual({ command: 'reset' });
  });

  it('returns error for unknown command', () => {
    const result = parseTrickCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('delegates to extra parser', () => {
    const result = parseTrickCommand('bid 3', ['bid'], (cmd, args) => {
      if (cmd === 'bid') return { command: 'bid', bid: Number(args[0]) };
      return null;
    });
    expect(result).toEqual({ command: 'bid', bid: 3 });
  });

  it('extra parser error is returned', () => {
    const result = parseTrickCommand('bid', ['bid'], () => {
      return { error: 'custom error' };
    });
    expect('error' in result).toBe(true);
  });
});
