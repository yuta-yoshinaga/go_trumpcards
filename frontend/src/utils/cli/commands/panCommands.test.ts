import { describe, expect, it } from 'vitest';
import { parsePanCommand } from './panCommands';

describe('parsePanCommand', () => {
  it('parses drawstock', () => {
    expect(parsePanCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parsePanCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parsePanCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parsePanCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses meld with 3+ indices', () => {
    expect(parsePanCommand('m 0 1 2')).toEqual({ args: ['meld', { cardIndices: [0, 1, 2] }] });
    expect(parsePanCommand('meld 3 4 5 6')).toEqual({ args: ['meld', { cardIndices: [3, 4, 5, 6] }] });
  });

  it('returns error for meld with fewer than 3 indices', () => {
    expect('error' in parsePanCommand('m 0 1')).toBe(true);
  });

  it('returns error for meld with a non-numeric index', () => {
    expect('error' in parsePanCommand('m 0 1 x')).toBe(true);
  });

  it('parses layoff with owner, meld, and card index', () => {
    expect(parsePanCommand('lo 1 0 3')).toEqual({
      args: ['layoff', { meldOwner: 1, meldIdx: 0, cardIndex: 3 }],
    });
    expect(parsePanCommand('layoff 2 1 4')).toEqual({
      args: ['layoff', { meldOwner: 2, meldIdx: 1, cardIndex: 4 }],
    });
  });

  it('returns error for layoff missing arguments', () => {
    expect('error' in parsePanCommand('lo 1 0')).toBe(true);
  });

  it('parses discard with index', () => {
    expect(parsePanCommand('d 3')).toEqual({ args: ['discard', { cardIndex: 3 }] });
    expect(parsePanCommand('discard 5')).toEqual({ args: ['discard', { cardIndex: 5 }] });
  });

  it('returns error for discard without index', () => {
    expect('error' in parsePanCommand('d')).toBe(true);
  });

  it('parses nextround', () => {
    expect(parsePanCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parsePanCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parsePanCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parsePanCommand('r')).toEqual({ args: ['reset'] });
    expect(parsePanCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    expect('error' in parsePanCommand('xyz')).toBe(true);
  });

  it('suggests a near command', () => {
    const result = parsePanCommand('discrd');
    expect('error' in result).toBe(true);
  });
});
