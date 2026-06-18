import { describe, expect, it } from 'vitest';
import { parseKalookiCommand } from './kalookiCommands';

describe('parseKalookiCommand', () => {
  it('parses drawstock', () => {
    expect(parseKalookiCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseKalookiCommand('drawstock')).toEqual({ args: ['drawstock'] });
  });

  it('parses drawdiscard', () => {
    expect(parseKalookiCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseKalookiCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  it('parses meld with indices into one meld group', () => {
    expect(parseKalookiCommand('meld 0 1 2')).toEqual({ args: ['meld', { meldGroups: [[0, 1, 2]] }] });
    expect(parseKalookiCommand('m 3 4 5')).toEqual({ args: ['meld', { meldGroups: [[3, 4, 5]] }] });
  });

  it('parses meld without indices as empty group', () => {
    expect(parseKalookiCommand('meld')).toEqual({ args: ['meld', { meldGroups: [[]] }] });
  });

  it('parses layoff with three indices', () => {
    expect(parseKalookiCommand('lo 1 0 4')).toEqual({
      args: ['layoff', { targetPlayerIdx: 1, meldIdx: 0, cardIndex: 4 }],
    });
    expect(parseKalookiCommand('layoff 2 3 5')).toEqual({
      args: ['layoff', { targetPlayerIdx: 2, meldIdx: 3, cardIndex: 5 }],
    });
  });

  it('returns usage error for layoff with missing indices', () => {
    expect('error' in parseKalookiCommand('lo')).toBe(true);
    expect('error' in parseKalookiCommand('lo 1')).toBe(true);
    expect('error' in parseKalookiCommand('lo 1 0')).toBe(true);
  });

  it('returns usage error for layoff with invalid/negative indices', () => {
    expect('error' in parseKalookiCommand('lo a b c')).toBe(true);
    expect('error' in parseKalookiCommand('lo -1 0 4')).toBe(true);
  });

  it('parses discard with index', () => {
    expect(parseKalookiCommand('d 3')).toEqual({ args: ['discard', { cardIndex: 3 }] });
    expect(parseKalookiCommand('dis 5')).toEqual({ args: ['discard', { cardIndex: 5 }] });
    expect(parseKalookiCommand('discard 1')).toEqual({ args: ['discard', { cardIndex: 1 }] });
  });

  it('returns error for discard without index', () => {
    const result = parseKalookiCommand('d');
    expect('error' in result).toBe(true);
  });

  it('parses nextround', () => {
    expect(parseKalookiCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseKalookiCommand('nextround')).toEqual({ args: ['nextround'] });
  });

  it('parses log', () => {
    expect(parseKalookiCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseKalookiCommand('r')).toEqual({ args: ['reset'] });
    expect(parseKalookiCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error with suggestion for unknown command', () => {
    const result = parseKalookiCommand('xyz');
    expect('error' in result).toBe(true);
  });

  it('suggests a near-miss command', () => {
    const result = parseKalookiCommand('rese');
    expect('error' in result && result.error).toContain('Did you mean');
  });
});
