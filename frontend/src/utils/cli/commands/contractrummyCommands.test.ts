import { describe, expect, it } from 'vitest';
import { parseContractRummyCommand } from './contractrummyCommands';

describe('parseContractRummyCommand', () => {
  it('parses draw commands', () => {
    expect(parseContractRummyCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseContractRummyCommand('drawstock')).toEqual({ args: ['drawstock'] });
    expect(parseContractRummyCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseContractRummyCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses discard with a single index', () => {
    expect(parseContractRummyCommand('discard 3')).toEqual({ args: ['discard', { cardIndex: 3 }] });
    expect(parseContractRummyCommand('x 0')).toEqual({ args: ['discard', { cardIndex: 0 }] });
  });

  it('errors on discard without exactly one index', () => {
    expect('error' in parseContractRummyCommand('discard')).toBe(true);
    expect('error' in parseContractRummyCommand('discard 1 2')).toBe(true);
    expect('error' in parseContractRummyCommand('discard x')).toBe(true);
  });

  it('parses a multi-slot contract meld', () => {
    expect(parseContractRummyCommand('meld 0,1,2 / 3,4,5')).toEqual({
      args: [
        'meldcontract',
        {
          indicesPerSlot: [
            [0, 1, 2],
            [3, 4, 5],
          ],
        },
      ],
    });
  });

  it('accepts space-separated indices within a meld slot', () => {
    expect(parseContractRummyCommand('meld 0 1 2')).toEqual({
      args: ['meldcontract', { indicesPerSlot: [[0, 1, 2]] }],
    });
  });

  it('errors on an empty or invalid meld', () => {
    expect('error' in parseContractRummyCommand('meld')).toBe(true);
    expect('error' in parseContractRummyCommand('meld a,b')).toBe(true);
  });

  it('parses an extra meld with at least three cards', () => {
    expect(parseContractRummyCommand('extra 1,2,3')).toEqual({
      args: ['meldextra', { cardIndices: [1, 2, 3] }],
    });
  });

  it('errors on an extra meld with fewer than three cards', () => {
    expect('error' in parseContractRummyCommand('extra 1,2')).toBe(true);
  });

  it('parses layoff with card, player, and meld indices', () => {
    expect(parseContractRummyCommand('layoff 2 1 0')).toEqual({
      args: ['layoff', { cardIndex: 2, targetPlayerIdx: 1, meldIdx: 0 }],
    });
    expect(parseContractRummyCommand('lo 0 0 0')).toEqual({
      args: ['layoff', { cardIndex: 0, targetPlayerIdx: 0, meldIdx: 0 }],
    });
  });

  it('errors on layoff without three indices', () => {
    expect('error' in parseContractRummyCommand('layoff 1 2')).toBe(true);
  });

  it('parses nextround, log, and reset', () => {
    expect(parseContractRummyCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseContractRummyCommand('nextround')).toEqual({ args: ['nextround'] });
    expect(parseContractRummyCommand('log')).toEqual({ args: ['log'] });
    expect(parseContractRummyCommand('l')).toEqual({ args: ['log'] });
    expect(parseContractRummyCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseContractRummyCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a command for a close typo', () => {
    const result = parseContractRummyCommand('discrd 1');
    expect('error' in result).toBe(true);
    if ('error' in result) expect(result.error).toContain('discard');
  });

  it('errors on an unknown command', () => {
    expect('error' in parseContractRummyCommand('zzz')).toBe(true);
  });
});
