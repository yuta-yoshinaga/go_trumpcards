import { describe, expect, it } from 'vitest';
import { parseUltimateTexasHoldemCommand as parse } from './ultimatetexasholdemCommands';

describe('parseUltimateTexasHoldemCommand', () => {
  it('maps GUI bet, play, check, fold and reset actions to exec arguments', () => {
    expect(parse('bet 100 25')).toEqual({ args: ['bet', 100, 25] });
    expect(parse('b 50')).toEqual({ args: ['bet', 50, 0] });
    for (const mult of [1, 2, 3, 4])
      expect(parse(`play ${mult}`)).toEqual({ args: ['play', undefined, undefined, mult] });
    for (const mult of [1, 2, 3, 4])
      expect(parse(String(mult))).toEqual({ args: ['play', undefined, undefined, mult] });
    expect(parse('c')).toEqual({ args: ['check'] });
    expect(parse('check')).toEqual({ args: ['check'] });
    expect(parse('f')).toEqual({ args: ['fold'] });
    expect(parse('fold')).toEqual({ args: ['fold'] });
    expect(parse('r')).toEqual({ args: ['reset'] });
    expect(parse('reset')).toEqual({ args: ['reset'] });
  });
  it('rejects missing, malformed and out of range arguments', () => {
    for (const cmd of ['bet', 'bet 0', 'bet 4 -1', 'bet x', 'play', 'play 0', 'play 5', 'play x'])
      expect(parse(cmd)).toHaveProperty('error');
  });
  it('rejects unsupported API commands and unknown commands', () => {
    expect(parse('hint')).toHaveProperty('error');
    expect(parse('log')).toHaveProperty('error');
    expect(parse('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });
});
