import { describe, expect, it } from 'vitest';
import { parseConquianCommand } from './conquianCommands';

describe('parseConquianCommand', () => {
  it('parses drawstock', () => {
    expect(parseConquianCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseConquianCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parseConquianCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseConquianCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses meld with indices', () => {
    expect(parseConquianCommand('meld 0 1 2')).toEqual({ args: ['meld', undefined, undefined, [[0, 1, 2]]] });
    expect(parseConquianCommand('m 3 4 5')).toEqual({ args: ['meld', undefined, undefined, [[3, 4, 5]]] });
  });

  it('parses meld without indices as empty group', () => {
    expect(parseConquianCommand('meld')).toEqual({ args: ['meld', undefined, undefined, [[]]] });
  });

  it('parses discard with index', () => {
    expect(parseConquianCommand('d 3')).toEqual({ args: ['discard', 3] });
    expect(parseConquianCommand('dis 5')).toEqual({ args: ['discard', 5] });
    expect(parseConquianCommand('discard 1')).toEqual({ args: ['discard', 1] });
  });

  it('returns error for discard without index', () => {
    const result = parseConquianCommand('d');
    expect('error' in result).toBe(true);
  });

  it('parses nextround', () => {
    expect(parseConquianCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseConquianCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseConquianCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseConquianCommand('r')).toEqual({ args: ['reset'] });
    expect(parseConquianCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseConquianCommand('xyz');
    expect('error' in result).toBe(true);
  });
});
