import { describe, expect, it } from 'vitest';
import { parseCoincheCommand } from './coincheCommands';

describe('parseCoincheCommand', () => {
  it('maps every exposed action to its endpoint command and positional arguments', () => {
    const buttonCommands: Record<string, unknown[]> = {
      'bid 100 2': ['bid', 100, 2],
      'b 100 2': ['bid', 100, 2],
      'p 3': ['play', undefined, 3],
      'play 3': ['play', undefined, 3],
      pass: ['pass'],
      coinche: ['coinche'],
      c: ['coinche'],
      surcoinche: ['surcoinche'],
      sc: ['surcoinche'],
      decline: ['decline'],
      d: ['decline'],
      next: ['next'],
      n: ['next'],
      nextround: ['nextround'],
      nr: ['nextround'],
      hint: ['hint'],
      h: ['hint'],
      reset: ['reset'],
    };
    for (const [input, expected] of Object.entries(buttonCommands))
      expect(parseCoincheCommand(input)).toEqual({ args: expected });
    expect(parseCoincheCommand('r')).toEqual({ args: ['reset'] });
  });
  it('accepts bid and play boundary values, including Capot', () => {
    for (const input of ['bid 80 1', 'bid 160 4', 'bid 250 1', 'p 0', 'p 7'])
      expect(parseCoincheCommand(input)).toHaveProperty('args');
  });
  it('rejects missing and out of range arguments', () => {
    for (const input of ['bid', 'bid 100', 'bid 85 2', 'bid 100 0', 'bid 100 5', 'p', 'p -1', 'p 8'])
      expect(parseCoincheCommand(input)).toHaveProperty('error');
  });
  it('rejects commands absent from CoincheCommand', () => {
    expect(parseCoincheCommand('log')).toHaveProperty('error');
  });
});
