import { describe, expect, it } from 'vitest';
import { parseBrusquembilleCommand } from './brusquembilleCommands';

describe('parseBrusquembilleCommand', () => {
  it('matches play buttons and all API actions', () => {
    expect(parseBrusquembilleCommand('p 2')).toEqual({ args: ['play', 2] });
    expect(parseBrusquembilleCommand('play 0')).toEqual({ args: ['play', 0] });
    expect(parseBrusquembilleCommand('n')).toEqual({ args: ['next'] });
    expect(parseBrusquembilleCommand('next')).toEqual({ args: ['next'] });
    expect(parseBrusquembilleCommand('h')).toEqual({ args: ['hint'] });
    expect(parseBrusquembilleCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseBrusquembilleCommand('r')).toEqual({ args: ['reset'] });
    expect(parseBrusquembilleCommand('reset')).toEqual({ args: ['reset'] });
  });
  it('rejects missing, invalid, and negative card indexes', () => {
    for (const input of ['p', 'play x', 'p -1'])
      expect(parseBrusquembilleCommand(input)).toEqual({ error: 'Usage: p <idx>' });
  });
  it('rejects unknown commands', () => {
    expect(parseBrusquembilleCommand('bid')).toEqual({ error: 'Unknown command: bid' });
    expect(parseBrusquembilleCommand('l')).toEqual({ error: 'Unknown command: l. Did you mean: p?' });
    expect(parseBrusquembilleCommand('log')).toEqual({ error: 'Unknown command: log' });
  });
});
