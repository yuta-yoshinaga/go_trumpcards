import { describe, expect, it } from 'vitest';
import { parseBeloteCommand } from './beloteCommands';

describe('parseBeloteCommand', () => {
  it('maps card play to the API cardIndex slot used by the page action', () => {
    expect(parseBeloteCommand('p 0')).toEqual({ args: ['play', undefined, 0] });
    expect(parseBeloteCommand('play 7')).toEqual({ args: ['play', undefined, 7] });
  });
  it('rejects missing, nonnumeric, and out-of-range card indices', () => {
    for (const input of ['p', 'play x', 'p -1', 'p 8'])
      expect(parseBeloteCommand(input)).toEqual({ error: 'Usage: p <idx>' });
  });
  it('maps bidding actions to the commands used by the page buttons', () => {
    expect(parseBeloteCommand('orderup')).toEqual({ args: ['orderup', undefined] });
    expect(parseBeloteCommand('pass')).toEqual({ args: ['pass', undefined] });
    expect(parseBeloteCommand('calltrump 3')).toEqual({ args: ['calltrump', undefined, 3] });
  });
  it('rejects missing and out-of-range trump suit', () => {
    for (const input of ['calltrump', 'calltrump x', 'calltrump 0', 'calltrump 5']) {
      expect(parseBeloteCommand(input)).toEqual({ error: 'Usage: calltrump <1-4>' });
    }
  });
  it('maps advancement, hint, and reset aliases to supported API commands', () => {
    expect(parseBeloteCommand('n')).toEqual({ args: ['next', undefined] });
    expect(parseBeloteCommand('next')).toEqual({ args: ['next', undefined] });
    expect(parseBeloteCommand('nr')).toEqual({ args: ['nextround', undefined] });
    expect(parseBeloteCommand('nextround')).toEqual({ args: ['nextround', undefined] });
    expect(parseBeloteCommand('h')).toEqual({ args: ['hint', undefined] });
    expect(parseBeloteCommand('hint')).toEqual({ args: ['hint', undefined] });
    expect(parseBeloteCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBeloteCommand('reset')).toEqual({ args: ['reset'] });
  });
  it('rejects unsupported and unknown commands', () => {
    expect(parseBeloteCommand('bid 2')).toEqual({ error: expect.stringContaining('Unknown command: bid') });
    expect(parseBeloteCommand('wat')).toEqual({ error: 'Unknown command: wat' });
  });
});
