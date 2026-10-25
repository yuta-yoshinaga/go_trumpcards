import { describe, expect, it } from 'vitest';
import { parseRummy500Command } from './rummy500Commands';

describe('parseRummy500Command', () => {
  it('matches the API calls made by game actions', () => {
    expect(parseRummy500Command('d')).toEqual({ args: ['drawstock'] });
    expect(parseRummy500Command('take 2')).toEqual({ args: ['drawdiscard', undefined, undefined, undefined, 2] });
    expect(parseRummy500Command('meld 0 2 4')).toEqual({ args: ['meld', undefined, undefined, [0, 2, 4]] });
    expect(parseRummy500Command('discard 1')).toEqual({ args: ['discard', 1] });
    expect(parseRummy500Command('layoff 3 1 0')).toEqual({
      args: ['layoff', undefined, undefined, undefined, undefined, { cardIndex: 3, meldOwner: 1, meldIdx: 0 }],
    });
    expect(parseRummy500Command('nr')).toEqual({ args: ['nextround'] });
    expect(parseRummy500Command('r')).toEqual({ args: ['reset'] });
  });

  it('rejects missing, malformed, and out of range arguments', () => {
    for (const command of ['take', 'take -1', 'take x']) expect(parseRummy500Command(command)).toHaveProperty('error');
    for (const command of ['meld 0 1', 'meld 0 x 2', 'meld -1 2 3'])
      expect(parseRummy500Command(command)).toHaveProperty('error');
    for (const command of ['discard', 'discard -1', 'discard x'])
      expect(parseRummy500Command(command)).toHaveProperty('error');
    for (const command of ['layoff 0 1', 'layoff 0 -1 0', 'layoff x 1 0'])
      expect(parseRummy500Command(command)).toHaveProperty('error');
  });

  it('rejects unsupported commands including hint and log', () => {
    for (const command of ['hint', 'log', 'bogus']) expect(parseRummy500Command(command)).toHaveProperty('error');
  });
});
