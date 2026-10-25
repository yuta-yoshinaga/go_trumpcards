import { describe, expect, it } from 'vitest';
import { parsePutCommand } from './putCommands';

describe('parsePutCommand', () => {
  it('maps all page actions and aliases to putApi.exec arguments', () => {
    expect(parsePutCommand('p 0')).toEqual({ args: ['play', 0] });
    expect(parsePutCommand('play 2')).toEqual({ args: ['play', 2] });
    expect(parsePutCommand('put')).toEqual({ args: ['put'] });
    expect(parsePutCommand('accept')).toEqual({ args: ['accept'] });
    expect(parsePutCommand('decline')).toEqual({ args: ['decline'] });
    expect(parsePutCommand('n')).toEqual({ args: ['next'] });
    expect(parsePutCommand('next')).toEqual({ args: ['next'] });
    expect(parsePutCommand('h')).toEqual({ args: ['hint'] });
    expect(parsePutCommand('hint')).toEqual({ args: ['hint'] });
    expect(parsePutCommand('r')).toEqual({ args: ['reset'] });
    expect(parsePutCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('rejects missing, nonnumeric, and out-of-range card indices', () => {
    for (const input of ['p', 'play x', 'p 2x', 'p -1', 'p 3']) {
      expect(parsePutCommand(input)).toEqual({ error: 'Usage: p <idx> (0-2)' });
    }
  });

  it('rejects commands not supported by Put', () => {
    expect(parsePutCommand('bid 2')).toEqual({ error: 'Unknown command: bid' });
    expect(parsePutCommand('log')).toEqual({ error: 'Unknown command: log' });
  });
});
